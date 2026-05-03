package dnsbl

import (
	"sync"
	"time"
)

// slidingWindow limits how many events may occur within a trailing time window.
type slidingWindow struct {
	mu     sync.Mutex
	window time.Duration
	max    int
	events []time.Time
}

func (s *slidingWindow) allow(now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	cutoff := now.Add(-s.window)
	i := 0
	for i < len(s.events) && s.events[i].Before(cutoff) {
		i++
	}
	s.events = s.events[i:]
	if len(s.events) >= s.max {
		return false
	}
	s.events = append(s.events, now)
	return true
}

// minuteCounter resets each UTC minute (approximate rate per calendar minute).
type minuteCounter struct {
	mu      sync.Mutex
	minute  int64 // unix / 60
	count   int
	perMinute int
}

func (m *minuteCounter) allow(now time.Time) bool {
	if m.perMinute <= 0 {
		return true
	}
	slice := now.Unix() / 60
	m.mu.Lock()
	defer m.mu.Unlock()
	if slice != m.minute {
		m.minute = slice
		m.count = 0
	}
	if m.count >= m.perMinute {
		return false
	}
	m.count++
	return true
}

type clientWindows struct {
	mu      sync.Mutex
	clients map[string]*slidingWindow
	window  time.Duration
	max     int
}

func (c *clientWindows) allow(key string, now time.Time) bool {
	if c.max <= 0 {
		return true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.clients == nil {
		c.clients = make(map[string]*slidingWindow)
	}
	w := c.clients[key]
	if w == nil {
		w = &slidingWindow{window: c.window, max: c.max}
		c.clients[key] = w
	}
	ok := w.allow(now)
	if maxClientRLKeys > 0 && len(c.clients) > maxClientRLKeys {
		c.pruneHalfLocked()
	}
	return ok
}

func (c *clientWindows) pruneHalfLocked() {
	// Drop arbitrary half of client entries to bound memory under abuse.
	n := len(c.clients) / 2
	for k := range c.clients {
		delete(c.clients, k)
		n--
		if n <= 0 {
			break
		}
	}
}
