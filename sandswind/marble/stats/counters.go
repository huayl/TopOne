package stats

import (
	"bytes"
	"fmt"
	"sync"
)

// Counters
type Counters struct {
	mu     sync.Mutex
	counts map[string]int64
}

// NewCounters
func NewCounters(name string) *Counters {
	c := &Counters{counts: make(map[string]int64)}
	return c
}

// String
func (c *Counters) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return counterToString(c.counts)
}

// Add
func (c *Counters) Add(name string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[name] += value
}

// Set
func (c *Counters) Set(name string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[name] = value
}

// Counts
func (c *Counters) Counts() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()

	counts := make(map[string]int64, len(c.counts))
	for k, v := range c.counts {
		counts[k] = v
	}
	return counts
}

// CountersFunc
type CountersFunc func() map[string]int64

// Counts
func (f CountersFunc) Counts() map[string]int64 {
	return f()
}

// String.
func (f CountersFunc) String() string {
	m := f()
	if m == nil {
		return "{}"
	}
	return counterToString(m)
}

func counterToString(m map[string]int64) string {
	b := bytes.NewBuffer(make([]byte, 0, 4096))
	fmt.Fprintf(b, "{")
	firstValue := true
	for k, v := range m {
		if firstValue {
			firstValue = false
		} else {
			fmt.Fprintf(b, ", ")
		}
		fmt.Fprintf(b, "\"%v\": %v", k, v)
	}
	fmt.Fprintf(b, "}")
	return b.String()
}
