package ring

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
	"sync"

	node "toy_dynamodb/node"
)

type getResponse struct {
	nodeName string
	value    string
	ok       bool
}

const VirtualSpotCount = 100

type argError struct {
	arg     string
	message string
}

type quorumWriteError struct {
	message string
	w       int
	n       int
}
type quorumReadError struct {
	message string
	r       int
	n       int
}

func (e *quorumWriteError) Error() string {
	return fmt.Sprintf("%s ", e.message)
}

func (e *quorumReadError) Error() string {
	return fmt.Sprintf("%s ", e.message)
}
func (e *argError) Error() string {
	return fmt.Sprintf("%s - %s", e.arg, e.message)
}

type Ring struct {
	nodes        map[string]*node.Node
	sortedNodes  []uint64
	nodeMap      map[uint64]string
	rwmu         *sync.RWMutex
	ReplicaCount uint
}

func (r *Ring) AddNode(name string) error {

	r.rwmu.RLock()
	if r.nodes == nil {
		return &argError{fmt.Sprint(r.nodes), "Is Not initilized"}
	}
	_, exist := r.nodes[name]

	if exist {
		r.rwmu.RUnlock()
		return &argError{name, "Already Exist In Node"}
	}
	r.rwmu.RUnlock()

	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	// This double check is for fixing TOCTOU
	// This occurs while routine a runlocks and in exact that moment
	// routine b locks and adds to the ring, routine a thinks node doesn't exist
	// bun infact it do exist
	_, exist = r.nodes[name]

	if exist {
		return &argError{name, "Already Exist In Node"}
	}

	tNode := node.New(name)

	for i := range VirtualSpotCount {

		uintval := getHash(fmt.Sprintf("%s#%d", name, i))
		r.nodeMap[uintval] = name
		r.sortedNodes = append(r.sortedNodes, uintval)
	}
	slices.Sort(r.sortedNodes)

	r.nodes[name] = tNode
	return nil

}

func (r *Ring) Put(key, val string, w int) error {

	if len(r.nodes) < w {
		return &argError{fmt.Sprintf("%v count is %d", r.nodes, len(r.nodes)), "Write Quorum Count can't be greater than node counts"}
	}

	getNodes := r.getNode(key, int(r.ReplicaCount))

	if len(getNodes) == 0 {
		return &argError{fmt.Sprint(getNodes), " returned count 0"}
	}

	ch := make(chan error, len(getNodes))

	for _, name := range getNodes {

		go func(n string) {
			// for golang 1.25.6 loop variables are passed by ref not by value
			// but for good sake i'll stuck to old version
			err := r.nodes[n].Put(key, val)

			if err != nil {
				ch <- err
			} else {
				ch <- nil
			}

		}(name)
	}

	for s, f := 0, 0; ; {
		err := <-ch

		if err == nil {
			s++
		} else {
			f++
		}

		if s == w {
			return nil
		} else if s+f == len(getNodes) {
			return &quorumWriteError{message: "Failed to hit quorum", w: w, n: len(getNodes)}
		}
	}
}

func (r *Ring) Get(key string, q int) (map[string]string, error) {

	if len(r.nodes) < q {
		return nil, &argError{fmt.Sprintf("%v count is %d", r.nodes, len(r.nodes)), "Write Quorum Count can't be greater than node counts"}
	}

	getNodes := r.getNode(key, int(r.ReplicaCount))

	if len(getNodes) == 0 {
		return nil, &argError{fmt.Sprint(getNodes), " returned count 0"}
	}
	ch := make(chan getResponse, len(getNodes))

	vals := make(map[string]string)

	for _, name := range getNodes {

		go func(n string) {
			v, ok := r.nodes[n].Get(key)
			ch <- getResponse{
				nodeName: n,
				value:    v,
				ok:       ok,
			}
		}(name)
	}

	for s, f := 0, 0; ; {
		res := <-ch

		if res.ok {
			vals[res.nodeName] = res.value
			s++
		} else {
			f++
		}

		if s == q {
			return vals, nil
		} else if s+f == len(getNodes) {
			return nil, &quorumReadError{message: "Failed to hit quorum", r: q, n: len(getNodes)}
		}
	}

}

func (r *Ring) Init() {
	r.nodeMap = make(map[uint64]string)
	r.nodes = make(map[string]*node.Node)
	r.sortedNodes = []uint64{}
	r.rwmu = &sync.RWMutex{}
}

func (r *Ring) getNode(val string, n int) []string {

	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	seenSet := map[string]bool{}
	nodes := []string{}
	uintval := getHash(val)
	index := sort.Search(len(r.sortedNodes), func(i int) bool { return r.sortedNodes[i] >= uintval })

	for i, ct, steps := ((index) % len(r.sortedNodes)), 0, 0; steps < len(r.sortedNodes) && ct < n; i, steps = (i+1)%len(r.sortedNodes), steps+1 {

		if seenSet[r.nodeMap[r.sortedNodes[uint64(i)]]] != true {

			seenSet[r.nodeMap[r.sortedNodes[uint64(i)]]] = true
			nodes = append(nodes, r.nodeMap[r.sortedNodes[uint64(i)]])
			ct += 1
		}
	}

	return nodes

}

func getHash(val string) uint64 {
	shaSource := sha1.New()
	shaSource.Write([]byte(val))
	sum := shaSource.Sum(nil)

	return binary.BigEndian.Uint64(sum[:8])

}
