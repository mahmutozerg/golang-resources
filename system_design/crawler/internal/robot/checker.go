package checker

import (
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"
)

type FetcherFunc func(url string) ([]byte, error)

type Checker struct {
	hostRules map[string]*robotstxt.Group
	rwMu      sync.RWMutex
	AgentName string
	fetcher   FetcherFunc
}

func New(agentName string, fetcher FetcherFunc) *Checker {
	return &Checker{
		hostRules: make(map[string]*robotstxt.Group),
		AgentName: agentName,
		fetcher:   fetcher,
	}
}

func (c *Checker) IsAllowed(targetUrl *url.URL) bool {
	host := targetUrl.Host

	c.rwMu.RLock()
	group, ok := c.hostRules[host]
	c.rwMu.RUnlock()

	if ok {
		if group == nil {
			return true
		}
		return group.Test(targetUrl.Path)
	}

	c.rwMu.Lock()
	defer c.rwMu.Unlock()

	if group, ok = c.hostRules[host]; ok {
		if group == nil {
			return true
		}
		return group.Test(targetUrl.Path)
	}

	robotsURL := targetUrl.Scheme + "://" + targetUrl.Host + "/robots.txt"
	data, err := c.fetcher(robotsURL)

	// Failed to get robots.txt
	if err != nil {

		return false
	}

	// Empty or nill robots.txt, allow, rules nil means you can visit everywhere
	if len(data) == 0 {

		c.hostRules[host] = nil
		return true
	}

	// Failed to parse, assume disallowed
	rules, err := robotstxt.FromBytes(data)
	if err != nil {
		return false
	}

	group = rules.FindGroup(c.AgentName)
	c.hostRules[host] = group

	if group == nil {
		return true
	}
	return group.Test(targetUrl.Path)
}
