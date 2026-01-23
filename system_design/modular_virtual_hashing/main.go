package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"time"
)

type Ring struct {
	sortedNodes []uint64
	nodeMap     map[uint64]string
}

const VirtualSpotCount = 100

func GetHash(val string) uint64 {
	shaSource := sha1.New()
	shaSource.Write([]byte(val))
	sum := shaSource.Sum(nil)

	return binary.BigEndian.Uint64(sum[:8])

}
func (r *Ring) AddNode(name string) {

	for i := range VirtualSpotCount {

		uintval := GetHash(fmt.Sprintf("%s#%d", name, i))
		r.nodeMap[uintval] = name
		r.sortedNodes = append(r.sortedNodes, uintval)
	}

	slices.Sort(r.sortedNodes)

}

func (r *Ring) RemoveNode(name string) {

	slices.DeleteFunc(r.sortedNodes, func(u uint64) bool {

		return r.nodeMap[u] == name

	})

	maps.DeleteFunc(r.nodeMap, func(k uint64, v string) bool {

		return v == name

	})

	slices.Sort(r.sortedNodes)

}
func (r Ring) GetNode(val string) string {

	uintval := GetHash(val)

	index := sort.Search(len(r.sortedNodes), func(i int) bool { return r.sortedNodes[i] >= uintval })

	if index >= len(r.sortedNodes) {
		return r.nodeMap[r.sortedNodes[uint64(0)]]
	}

	return r.nodeMap[r.sortedNodes[uint64(index)]]

}

func (r Ring) GetNodeByHash(val uint64) string {

	index := sort.Search(len(r.sortedNodes), func(i int) bool { return r.sortedNodes[i] >= val })

	if index >= len(r.sortedNodes) {
		return r.nodeMap[r.sortedNodes[uint64(0)]]
	}

	return r.nodeMap[r.sortedNodes[uint64(index)]]

}

func main() {

	r := Ring{}
	r.nodeMap = make(map[uint64]string)

	var hits = make(map[string]int)

	r.AddNode("node_1")
	r.AddNode("node_2")
	r.AddNode("node_3")
	r.AddNode("node_4")

	datas := func() []string {

		out := make([]string, 0, 1_000_000)
		for i := 0; i < 1_000_000; i++ {
			out = append(out, "data_"+strconv.FormatInt(int64(i), 10))
		}
		return out
	}()

	for i := range datas {

		val := r.GetNode(datas[i])
		hits[val]++
	}

	for i := range hits {
		fmt.Printf("%v count  %v\n", i, hits[i])
	}

	for k := range hits {
		delete(hits, k)
	}

	fmt.Println("*********Remove Node_1**********")
	start := time.Now()

	r.RemoveNode("node_1")

	for i := range datas {

		val := r.GetNode(datas[i])
		hits[val]++
	}

	for i := range hits {
		fmt.Printf("%v count  %v\n", i, hits[i])
	}

}
