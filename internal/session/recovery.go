package session

import (
	"context"

	netbridge "github.com/netbridge/netbridge"
)

type Recovery struct {
	state *State
}

func NewRecovery(state *State) *Recovery {
	return &Recovery{state: state}
}

func (r *Recovery) Recover(ctx context.Context) (*netbridge.Session, error) {
	session, err := r.state.Load()
	if err != nil {
		return nil, err
	}
	if session.Status == netbridge.StatusConnected {
		return session, nil
	}
	return nil, netbridge.ErrNoActiveSession
}
