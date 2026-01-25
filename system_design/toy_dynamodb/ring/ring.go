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

const VirtualSpotCount = 100

type argError struct {
	arg     string
	message string
}

func (e *argError) Error() string {
	return fmt.Sprintf("%s - %s", e.arg, e.message)
}

type Ring struct {
	nodes       map[string]*node.Node
	sortedNodes []uint64
	nodeMap     map[uint64]string
	rwmu        *sync.RWMutex
	RCount      uint
}

func (r *Ring) AddNode(name string) error {

	r.rwmu.RLock()
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

func (r *Ring) Put(key, val string) {

	getNodes := r.getNode(key, int(r.RCount))

	for _, name := range getNodes {

		r.nodes[name].Put(key, val)
	}

}

func (r *Ring) Get(key string) ([]string, bool) {

	getNodes := r.getNode(key, int(r.RCount))

	vals := []string{}

	for _, name := range getNodes {

		v, ok := r.nodes[name].Get(key)

		if !ok {
			continue
		}
		vals = append(vals, v)
	}

	return vals, len(vals) > 0
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
