package middleware

import (
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OriginCache struct {
	fetch func() ([]string, error)
	ttl   time.Duration

	mu        sync.RWMutex
	origins   map[string]struct{}
	expiresAt time.Time
}

func NewOriginCache(fetch func() ([]string, error), ttl time.Duration) *OriginCache {
	return &OriginCache{
		fetch:   fetch,
		ttl:     ttl,
		origins: map[string]struct{}{},
	}
}

func (c *OriginCache) IsAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	c.mu.RLock()
	if time.Now().Before(c.expiresAt) {
		_, ok := c.origins[origin]
		c.mu.RUnlock()
		return ok
	}
	c.mu.RUnlock()

	c.refresh()

	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.origins[origin]
	return ok
}

func (c *OriginCache) Invalidate() {
	c.mu.Lock()
	c.expiresAt = time.Time{}
	c.mu.Unlock()
}

func (c *OriginCache) refresh() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.expiresAt) {
		return
	}

	uris, err := c.fetch()
	if err != nil {
		log.Printf("[ERROR] OriginCache refresh failed: %v", err)
		c.expiresAt = time.Now().Add(5 * time.Second)
		return
	}

	origins := make(map[string]struct{}, len(uris))
	for _, uri := range uris {
		if origin := originFromURL(uri); origin != "" {
			origins[origin] = struct{}{}
		}
	}

	c.origins = origins
	c.expiresAt = time.Now().Add(c.ttl)
}

func originFromURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	scheme := strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Host)
	if (scheme == "https" && strings.HasSuffix(host, ":443")) ||
		(scheme == "http" && strings.HasSuffix(host, ":80")) {
		host = host[:strings.LastIndex(host, ":")]
	}

	return scheme + "://" + host
}
