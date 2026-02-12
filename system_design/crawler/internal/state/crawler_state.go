package crawler_state

import "sync"

type VisitedLinks struct {
	Visited map[string]bool
	Rwmu    *sync.RWMutex
}
