package core

import "sync"

type Event struct {
	Type    string
	Payload any
}

type EventHandler func(Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (eb *EventBus) Subscribe(eventType string, handler EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, h := range eb.handlers[event.Type] {
		h(event)
	}
}
