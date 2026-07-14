package health

import (
	"context"
	"sync"
	"time"

	"github.com/netbridge/netbridge/internal/profile"
)

type Watchdog struct {
	profileMgr *profile.Manager
	engine     *Engine
	interval   time.Duration
	threshold  int
	failureMap map[string]int
	mu         sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewWatchdog(pm *profile.Manager, engine *Engine) *Watchdog {
	return &Watchdog{
		profileMgr: pm,
		engine:     engine,
		threshold:  3,
		failureMap: make(map[string]int),
	}
}

func (w *Watchdog) Start(ctx context.Context, interval time.Duration, failoverFn func(string)) {
	w.ctx, w.cancel = context.WithCancel(ctx)
	w.interval = interval

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-w.ctx.Done():
				return
			case <-ticker.C:
				w.checkAll(failoverFn)
			}
		}
	}()
}

func (w *Watchdog) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Watchdog) checkAll(failoverFn func(string)) {
	profiles, err := w.profileMgr.List(w.ctx)
	if err != nil {
		return
	}

	for _, p := range profiles {
		result, err := w.engine.Check(w.ctx, p.ID)
		if err != nil || !result.Reachable {
			w.mu.Lock()
			w.failureMap[p.ID]++
			failures := w.failureMap[p.ID]
			w.mu.Unlock()

			if failures >= w.threshold && failoverFn != nil {
				failoverFn(p.ID)
			}
		} else {
			w.mu.Lock()
			w.failureMap[p.ID] = 0
			w.mu.Unlock()
		}
	}
}
