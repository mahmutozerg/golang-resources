package node

import "sync"

type Node struct {
	Name  string
	items map[string]string
	rwmu  *sync.RWMutex
}

func (n *Node) Put(key, val string) {
	n.rwmu.Lock()
	defer n.rwmu.Unlock()
	n.items[key] = val
}

func (n *Node) Get(key string) (string, bool) {

	n.rwmu.RLock()
	defer n.rwmu.RUnlock()

	v, exist := n.items[key]
	return v, exist
}

func New(name string) *Node {
	return &Node{Name: name, items: make(map[string]string), rwmu: &sync.RWMutex{}}
}
