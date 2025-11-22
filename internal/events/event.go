package events

import (
	"context"
	"fmt"
	"sync"
)

// Event represents a domain event
type Event interface {
	Type() string
}

// EventHandler handles events
type EventHandler func(ctx context.Context, event Event) error

// EventBus manages event publishing and subscription
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler)
	Unsubscribe(eventType string, handler EventHandler)
}

// InMemoryEventBus is a simple in-memory event bus implementation
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() EventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

// Publish publishes an event to all subscribed handlers
func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mu.RLock()
	handlers := bus.handlers[event.Type()]
	bus.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("event handling errors: %v", errs)
	}

	return nil
}

// Subscribe subscribes a handler to an event type
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}

// Unsubscribe removes a handler from an event type
func (bus *InMemoryEventBus) Unsubscribe(eventType string, handler EventHandler) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	handlers := bus.handlers[eventType]
	for i, h := range handlers {
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
			bus.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

