package srv

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Cache stores the results of successful SRV record queries.
type Cache interface {
	Get(name string) (host string, port uint16, err error)
}

type entry struct {
	srv    *net.SRV
	expire time.Duration
}

type cache struct {
	duration   time.Duration
	record     map[string]entry
	recordLock *sync.Mutex
}

// New returns a new cache.
func New(duration time.Duration) Cache {
	c := &cache{
		duration:   duration,
		record:     make(map[string]entry),
		recordLock: &sync.Mutex{},
	}
	go c.startGC(duration * 10)
	return c
}

func (c *cache) startGC(interval time.Duration) {
	for {
		c.flushExpired()
		time.Sleep(interval)
	}
}

func (c *cache) flushExpired() {
	c.recordLock.Lock()
	for k, v := range c.record {
		if v.expired() {
			delete(c.record, k)
		}
	}
	c.recordLock.Unlock()
}

func (c *cache) newEntry(srv *net.SRV) entry {
	return entry{
		srv:    srv,
		expire: time.Duration(time.Now().Unix()) + (c.duration / time.Second),
	}
}

// Returns the updated values
func (c *cache) update(name string) (host string, port uint16, err error) {
	_, addrs, err := net.LookupSRV("", "", name)
	if err != nil {
		return "", 0, fmt.Errorf("error updating SRV cache: %s", err)
	}

	srvEntry := addrs[0]
	host = srvEntry.Target
	port = srvEntry.Port
	c.recordLock.Lock()
	c.record[name] = c.newEntry(srvEntry)
	c.recordLock.Unlock()
	return host, port, nil
}

func (e *entry) expired() bool {
	return time.Duration(time.Now().Unix()) > e.expire
}

func (c *cache) Get(name string) (host string, port uint16, err error) {
	c.recordLock.Lock()
	v, ok := c.record[name]
	c.recordLock.Unlock()
	if ok && !v.expired() {
		return v.srv.Target, v.srv.Port, nil
	}
	return c.update(name)
}
