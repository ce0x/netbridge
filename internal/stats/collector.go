package stats

import (
	"sync"
	"time"

	netbridge "github.com/netbridge/netbridge"
)

type Collector struct {
	mu          sync.RWMutex
	bytesUp     int64
	bytesDown   int64
	startTime   time.Time
	lastBytesUp int64
	lastBytesDown int64
	lastCheck   time.Time
}

func NewCollector() *Collector {
	return &Collector{
		startTime: time.Now(),
		lastCheck: time.Now(),
	}
}

func (c *Collector) RecordUp(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bytesUp += n
}

func (c *Collector) RecordDown(n int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bytesDown += n
}

func (c *Collector) Snapshot() netbridge.TrafficStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	elapsed := now.Sub(c.lastCheck).Seconds()

	var rateUp, rateDown float64
	if elapsed > 0 {
		rateUp = float64(c.bytesUp-c.lastBytesUp) / elapsed
		rateDown = float64(c.bytesDown-c.lastBytesDown) / elapsed
	}

	return netbridge.TrafficStats{
		BytesUp:   c.bytesUp,
		BytesDown: c.bytesDown,
		RateUp:    rateUp,
		RateDown:  rateDown,
		Uptime:    now.Sub(c.startTime),
	}
}

func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bytesUp = 0
	c.bytesDown = 0
	c.startTime = time.Now()
	c.lastCheck = time.Now()
}
