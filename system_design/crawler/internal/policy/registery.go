package registery

import (
	"net/url"
	"sync"
	"time"

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
}

func New(agentName string, fetcher FetcherFunc) *Checker {
	return &Checker{
		hostRules: make(map[string]*DomainPolicy),
		AgentName: agentName,
		fetcher:   fetcher,
	}
}

func (c *Checker) GetPolicy(targetUrl *url.URL) (*DomainPolicy, error) {

	host := targetUrl.Host
	c.rwMu.RLock()
	existingPolicy, ok := c.hostRules[host]
	c.rwMu.RUnlock()
	if ok {
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

	c.hostRules[host] = newPolicy

	return newPolicy, nil
}
