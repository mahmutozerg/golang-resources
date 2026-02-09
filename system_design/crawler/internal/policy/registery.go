package registery

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mahmutozerg/golang-resources/system_design/crawler/internal/config"
	"github.com/temoto/robotstxt"
	"golang.org/x/time/rate"
)

type FetcherFunc func(url string) ([]byte, error)

type Checker struct {
	hostRules map[string]*DomainPolicy
	rwMu      sync.RWMutex
	AgentName string
	fetcher   FetcherFunc
}
type DomainPolicy struct {
	Rule    *robotstxt.Group
	Limiter *rate.Limiter
	Lru     atomic.Int64
}

func New(agentName string, fetcher FetcherFunc, ctx context.Context) *Checker {

	c := &Checker{
		hostRules: make(map[string]*DomainPolicy),
		AgentName: agentName,
		fetcher:   fetcher,
	}
	go func() {
		evictStale(c, ctx)
	}()

	return c
}

func (c *Checker) GetPolicy(targetUrl *url.URL) (*DomainPolicy, error) {

	host := targetUrl.Host
	c.rwMu.RLock()
	existingPolicy, ok := c.hostRules[host]
	c.rwMu.RUnlock()

	if ok {
		existingPolicy.Lru.Store(time.Now().UnixNano())
		return existingPolicy, nil
	}

	robotsURL := targetUrl.Scheme + "://" + targetUrl.Host + "/robots.txt"
	rBytes, err := c.fetcher(robotsURL)

	var rule *robotstxt.Group = nil
	var delay time.Duration = 1

	if err == nil && len(rBytes) > 0 {
		rData, errParse := robotstxt.FromBytes(rBytes)
		if errParse == nil {
			rule = rData.FindGroup(c.AgentName)
			if rule != nil {
				delay = rule.CrawlDelay
			}
		}
	}

	newPolicy := &DomainPolicy{
		Rule:    rule,
		Limiter: rate.NewLimiter(rate.Every(delay), 1),
	}

	c.rwMu.Lock()
	defer c.rwMu.Unlock()

	if existingPolicy, ok = c.hostRules[host]; ok {
		return existingPolicy, nil
	}

	newPolicy.Lru.Store(time.Now().UnixNano())
	c.hostRules[host] = newPolicy

	return newPolicy, nil
}

func evictStale(c *Checker, ctx context.Context) {

	d, _ := time.ParseDuration(config.EvictStaleTime)

	t := time.NewTicker(d)

loop:
	for {

		select {
		case <-t.C:
			c.rwMu.Lock()
			startLen := len(c.hostRules)
			fmt.Printf("Timer reached, evict stale active starting len %v", startLen)

			threshold := time.Now().Add(-d).UnixNano()

			maps.DeleteFunc(c.hostRules, func(k string, v *DomainPolicy) bool {
				return v.Lru.Load() < threshold
			})
			c.rwMu.Unlock()
			fmt.Printf("Cleanup completed,Cleaned links %v", startLen-len(c.hostRules))

		case <-ctx.Done():
			break loop
		}
	}

}
