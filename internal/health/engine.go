package health

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	netbridge "github.com/netbridge/netbridge"
	"github.com/netbridge/netbridge/internal/profile"
)

type Engine struct {
	profileMgr  *profile.Manager
	failureCb   func(string, *netbridge.HealthResult)
	watchdogCtx context.Context
	watchdogFn  context.CancelFunc
	mu          sync.Mutex
}

func NewEngine(pm *profile.Manager) *Engine {
	return &Engine{
		profileMgr: pm,
	}
}

func (e *Engine) Check(ctx context.Context, profileID string) (*netbridge.HealthResult, error) {
	p, err := e.profileMgr.Get(ctx, profileID)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", p.Server, p.Port), 10*time.Second)
	latency := time.Since(start)

	result := &netbridge.HealthResult{
		ProfileID:  profileID,
		Reachable:  err == nil,
		Latency:    latency,
		PacketLoss: 0,
		CheckedAt:  time.Now(),
	}

	if err != nil {
		result.Error = err.Error()
	}

	if conn != nil {
		conn.Close()
	}

	return result, nil
}

func (e *Engine) CheckAll(ctx context.Context) ([]*netbridge.HealthResult, error) {
	profiles, err := e.profileMgr.List(ctx)
	if err != nil {
		return nil, err
	}

	var results []*netbridge.HealthResult
	for _, p := range profiles {
		result, err := e.Check(ctx, p.ID)
		if err != nil {
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (e *Engine) StartWatchdog(ctx context.Context, interval time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.watchdogFn != nil {
		return fmt.Errorf("watchdog already running")
	}

	watchCtx, cancel := context.WithCancel(ctx)
	e.watchdogCtx = watchCtx
	e.watchdogFn = cancel

	go e.watchloop(watchCtx, interval)
	return nil
}

func (e *Engine) StopWatchdog() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.watchdogFn != nil {
		e.watchdogFn()
		e.watchdogFn = nil
	}
	return nil
}

func (e *Engine) OnFailure(fn func(profileID string, result *netbridge.HealthResult)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.failureCb = fn
}

func (e *Engine) watchloop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results, _ := e.CheckAll(ctx)
			for _, r := range results {
				if !r.Reachable && e.failureCb != nil {
					e.failureCb(r.ProfileID, r)
				}
			}
		}
	}
}
