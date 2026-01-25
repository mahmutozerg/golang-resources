package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"maps"
	"slices"
	"sort"
	"sync"
)

type Ring struct {
	sortedNodes []uint64
	nodeMap     map[uint64]string
	rwmu        *sync.RWMutex
}

const VirtualSpotCount = 100

func GetHash(val string) uint64 {
	shaSource := sha1.New()
	shaSource.Write([]byte(val))
	sum := shaSource.Sum(nil)

	return binary.BigEndian.Uint64(sum[:8])

}
func (r *Ring) AddNode(name string) {
	r.rwmu.Lock()
	defer r.rwmu.Unlock()
	for i := range VirtualSpotCount {

		uintval := GetHash(fmt.Sprintf("%s#%d", name, i))
		r.nodeMap[uintval] = name
		r.sortedNodes = append(r.sortedNodes, uintval)
	}

	slices.Sort(r.sortedNodes)

}

func (r *Ring) RemoveNode(name string) {

	r.rwmu.Lock()
	defer r.rwmu.Unlock()
	r.sortedNodes = slices.DeleteFunc(r.sortedNodes, func(u uint64) bool {

		return r.nodeMap[u] == name

	})

	maps.DeleteFunc(r.nodeMap, func(k uint64, v string) bool {

		return v == name

	})

	slices.Sort(r.sortedNodes)

}
func (r *Ring) GetNode(val string, n int) []string {

	r.rwmu.RLock()
	defer r.rwmu.RUnlock()

	seenSet := map[string]bool{}
	nodes := []string{}
	uintval := GetHash(val)
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

func main() {

	r := Ring{}
	r.nodeMap = make(map[uint64]string)
	r.rwmu = &sync.RWMutex{}
	var hits = make(map[string]int)

	r.AddNode("node_1")
	r.AddNode("node_2")
	r.AddNode("node_3")
	r.AddNode("node_4")

	datas := func() []string {

		out := make([]string, 0, 1_000_000)
		for i := range 1_000_000 {

			out = append(out, fmt.Sprintf("data_%d", i))
		}
		return out
	}()

	fmt.Println("1_000_000 is distributed in to 4 nodes with 3 replicas")

	for i := range datas {

		val := r.GetNode(datas[i], 3)

		for _, j := range val {
			hits[j]++

		}
	}

	for i := range hits {
		fmt.Printf("%v #  %v\n", i, hits[i])
	}

	r.RemoveNode("node_1")
	for k := range hits {
		delete(hits, k)
	}

	fmt.Println("*********NODE_1 DELETED************")
	fmt.Println("1_000_000 is distributed in to 3 nodes with 3 replicas")

	for i := range datas {

		val := r.GetNode(datas[i], 3)

		for _, j := range val {
			hits[j]++

		}
	}

	for i := range hits {
		fmt.Printf("%v #  %v\n", i, hits[i])
	}

}
