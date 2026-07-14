package core

import (
	"sync"

	netbridge "github.com/netbridge/netbridge"
)

type RuntimeState struct {
	mu            sync.RWMutex
	ActiveProfile *netbridge.Profile
	ActiveSession *netbridge.Session
	Status        netbridge.ConnectionStatus
}

func NewRuntimeState() *RuntimeState {
	return &RuntimeState{
		Status: netbridge.StatusDisconnected,
	}
}

func (rs *RuntimeState) GetStatus() netbridge.ConnectionStatus {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.Status
}

func (rs *RuntimeState) SetStatus(s netbridge.ConnectionStatus) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Status = s
}
