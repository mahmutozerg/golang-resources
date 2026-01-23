package main

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"slices"
	"sort"
)

type Ring struct {
	sortedNodes []uint64 //Sunucularımın adlarının hash değeri sortlanmış,
	nodeMap     map[uint64]string
}

func GetHash(val string) uint64 {
	shaSource := sha1.New()
	shaSource.Write([]byte(val))
	sum := shaSource.Sum(nil)

	return binary.BigEndian.Uint64(sum[:8])

}
func (r *Ring) AddNode(name string) {

	uintval := GetHash(name)
	r.nodeMap[uintval] = name

	r.sortedNodes = append(r.sortedNodes, uintval)
	fmt.Printf("New node with name %v  added with value %x\n", name, uintval)
	slices.Sort(r.sortedNodes)

}

func (r Ring) PrintSortedNodes() {

	for i := range r.sortedNodes {
		fmt.Printf("Node #%d %x\n", i, r.sortedNodes[i])
	}
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

	r.AddNode("node_1")
	r.AddNode("node_2")
	r.AddNode("node_3")

	r.PrintSortedNodes()

	r.AddNode("node_4")
	r.PrintSortedNodes()

	search := GetHash("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	fmt.Printf("%v found at  %s", search, r.GetNodeByHash(search))
}
