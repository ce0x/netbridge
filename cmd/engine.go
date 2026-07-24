package main

import (
	"sync"

	"github.com/netbridge/netbridge/internal/config"
	"github.com/netbridge/netbridge/internal/core"
)

var (
	engineOnce sync.Once
	engineInst *core.Engine
	engineErr  error
)

func getEngine() (*core.Engine, error) {
	engineOnce.Do(func() {
		cfg, err := config.Load()
		if err != nil {
			engineErr = err
			return
		}
		engineInst, engineErr = core.New(cfg)
	})
	return engineInst, engineErr
}
