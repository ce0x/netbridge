package core

import (
	"github.com/netbridge/netbridge/internal/config"
)

type DI struct {
	Engine *Engine
	Config *config.Config
}

func NewDI(cfg *config.Config) (*DI, error) {
	engine, err := New(cfg)
	if err != nil {
		return nil, err
	}
	return &DI{
		Engine: engine,
		Config: cfg,
	}, nil
}
