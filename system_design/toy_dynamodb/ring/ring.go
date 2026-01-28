package ring

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"sync"
	custom_errors "toy_dynamodb/Errors"
	kv "toy_dynamodb/proto"

	"github.com/cespare/xxhash/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const VirtualSpotCount = 100

type doOpReq struct {
	key, val string
	w        int
	isDelete bool
}

type getResponse struct {
	nodeName string
	value    string
	ok       bool
	err      error
}
type Ring struct {
	nodes        map[string]kv.KVStoreClient
	connections  []kv.KVStoreClient
	sortedNodes  []uint64
	nodeMap      map[uint64]string
	rwmu         *sync.RWMutex
	ReplicaCount uint
}

func (r *Ring) AddNode(address string) error {

	r.rwmu.RLock()
	if r.nodes == nil {
		return &custom_errors.ArgError{Arg: fmt.Sprint(r.nodes), Message: "Is Not initilized"}
	}
	_, exist := r.nodes[address]

	if exist {
		r.rwmu.RUnlock()
		return &custom_errors.ArgError{Arg: address, Message: "Already Exist In Node"}
	}
	r.rwmu.RUnlock()

	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	// This double check is for fixing TOCTOU
	// This occurs while routine a runlocks and in exact that moment
	// routine b locks and adds to the ring, routine a thinks node doesn't exist
	// bun infact it do exist
	_, exist = r.nodes[address]

	if exist {
		return &custom_errors.ArgError{Arg: address, Message: "Already Exist In Node"}
	}

	nodeConnection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	for i := range VirtualSpotCount {

		key := address + "#" + strconv.Itoa(i)
		uintval := getHash(key)
		r.nodeMap[uintval] = address
		r.sortedNodes = append(r.sortedNodes, uintval)
	}
	slices.Sort(r.sortedNodes)
	c := kv.NewKVStoreClient(nodeConnection)

	r.nodes[address] = c
	r.connections = append(r.connections, c)

	return nil

}

func (r *Ring) Get(key string, q int) (map[string]string, error) {

	if len(r.nodes) < q {
		return nil, &custom_errors.ArgError{Arg: fmt.Sprintf("%v count is %d", r.nodes, len(r.nodes)), Message: "Write Quorum Count can't be greater than node counts"}
	}

	getNodes := r.getNode(key, int(r.ReplicaCount))

	if len(getNodes) == 0 {
		return nil, &custom_errors.ArgError{Arg: fmt.Sprint(getNodes), Message: " returned count 0"}
	}
	vals := make(map[string]string)
	nodes := make([]struct {
		name string
		nd   kv.KVStoreClient
	}, 0, len(getNodes))

	r.rwmu.RLock()
	for _, n := range getNodes {
		nodes = append(nodes, struct {
			name string
			nd   kv.KVStoreClient
		}{n, r.nodes[n]})
	}
	r.rwmu.RUnlock()

	ch := make(chan getResponse, len(nodes))
	for _, p := range nodes {
		go func(n string, nd kv.KVStoreClient) {
			v, err := nd.Get(context.Background(), &kv.GetRequest{Key: key})
			if err != nil {
				ch <- getResponse{nodeName: n, err: err, ok: false}
				return
			}
			ch <- getResponse{nodeName: n, value: string(v.Value), ok: v.Found, err: err}
		}(p.name, p.nd)
	}

	s, f := 0, 0
	for {
		res := <-ch

		if res.ok {
			vals[res.nodeName] = res.value
			s++
		} else {
			f++
		}

		if s == q {
			return vals, nil
		} else if f == len(getNodes) {
			return nil, &custom_errors.QuorumReadError{Message: fmt.Sprintf("%s not found at any node", key), R: q, N: len(getNodes)}

		} else if s+f == len(getNodes) {

			return nil, &custom_errors.QuorumReadError{Message: "Failed to hit quorum", R: q, N: len(getNodes)}

		}
	}

}

func (r *Ring) Put(key, val string, w int) error {
	// pass by address for get rid unnecessary copies
	return r.doOp(&doOpReq{key: key, val: val, w: w, isDelete: false})
}

func (r *Ring) Delete(key string, w int) error {

	return r.doOp(&doOpReq{key: key, w: w, isDelete: true})
}

func (r *Ring) Init() {
	r.nodeMap = make(map[uint64]string)
	r.nodes = make(map[string]kv.KVStoreClient)
	r.sortedNodes = []uint64{}
	r.rwmu = &sync.RWMutex{}
}

// Burası ramde test yapabilmek için var olan bir yer genel logici test etiyoruz yani
func (r *Ring) RegisterClient(address string, client kv.KVStoreClient) {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()

	if _, exists := r.nodes[address]; exists {
		return
	}

	// 1. Client'ı kaydet
	r.nodes[address] = client
	r.connections = append(r.connections, client)

	for i := range VirtualSpotCount {
		key := address + "#" + strconv.Itoa(i)
		uintval := getHash(key)
		r.nodeMap[uintval] = address
		r.sortedNodes = append(r.sortedNodes, uintval)
	}
	slices.Sort(r.sortedNodes)
}

func (r *Ring) getNode(val string, n int) []string {

	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	seenSet := map[string]bool{}
	nodes := []string{}
	uintval := getHash(val)
	index := sort.Search(len(r.sortedNodes), func(i int) bool { return r.sortedNodes[i] >= uintval })

	for i, ct, steps := ((index) % len(r.sortedNodes)), 0, 0; steps < len(r.sortedNodes) && ct < n; i, steps = (i+1)%len(r.sortedNodes), steps+1 {

		if seenSet[r.nodeMap[r.sortedNodes[i]]] != true {

			seenSet[r.nodeMap[r.sortedNodes[i]]] = true
			nodes = append(nodes, r.nodeMap[r.sortedNodes[i]])
			ct += 1

		}
	}
	return nodes

}

func (r *Ring) doOp(rq *doOpReq) error {

	if len(r.nodes) < rq.w {
		return &custom_errors.ArgError{Arg: fmt.Sprintf("%v count is %d", r.nodes, len(r.nodes)), Message: "Write Quorum Count can't be greater than node counts"}
	}

	getNodes := r.getNode(rq.key, int(r.ReplicaCount))

	if len(getNodes) == 0 {
		return &custom_errors.ArgError{Arg: fmt.Sprint(getNodes), Message: " returned count 0"}
	}

	nodes := make([]kv.KVStoreClient, 0, len(getNodes))
	r.rwmu.RLock()

	for _, n := range getNodes {
		nd := r.nodes[n]
		if nd != nil {
			nodes = append(nodes, nd)
		}
	}
	r.rwmu.RUnlock()
	ch := make(chan error, len(nodes))

	for _, nd := range nodes {
		go func(nd kv.KVStoreClient) {
			var err error
			var deleteRes *kv.DeleteResponse
			var putRes *kv.PutResponse
			if rq.isDelete {
				deleteRes, err = nd.Delete(context.Background(), &kv.DeleteRequest{Key: rq.key})
				if err != nil { // Önce ağ hatası kontrolü
					ch <- err
					return
				}

				if !deleteRes.Success {
					err = fmt.Errorf("Something Went Wrong DeleteRes Returned false without error")
				}

			} else {
				putRes, err = nd.Put(context.Background(), &kv.PutRequest{Key: rq.key, Value: []byte(rq.val)})
				if err != nil { // Önce ağ hatası kontrolü
					ch <- err
					return
				}
				if !putRes.Success {
					err = fmt.Errorf("Something Went Wrong DeleteRes Returned false without error")
				}

			}
			ch <- err
		}(nd)
	}

	s, f := 0, 0

	for {
		err := <-ch

		if err == nil {
			s++
		} else {
			f++
		}

		if s == rq.w {
			return nil
		} else if s+f == len(nodes) {
			return &custom_errors.QuorumWriteError{Message: "Failed to hit quorum", W: rq.w, N: len(nodes)}
		}
	}
}

func getHash(val string) uint64 {
	return xxhash.Sum64String(val)
}
