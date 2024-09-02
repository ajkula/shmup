package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type EventManager struct {
	core.BaseSystem
	eventChan   chan interfaces.Event
	subscribers map[interfaces.EventType][]chan interfaces.Event
	mu          sync.RWMutex
}

func NewEventManager() interfaces.EventManagerInterface {
	return &EventManager{
		eventChan:   make(chan interfaces.Event, 5000),
		subscribers: make(map[interfaces.EventType][]chan interfaces.Event),
	}
}

func (em *EventManager) Initialize(ctx context.Context) error {
	if err := em.BaseSystem.Initialize(ctx); err != nil {
		return err
	}
	go em.processEvents()
	return nil
}

func (em *EventManager) processEvents() {
	for {
		select {
		case <-em.CTX.Done():
			return
		case event := <-em.eventChan:
			em.dispatch(event)
		}
	}
}

func (em *EventManager) Update(deltaTime float64) error {
	select {
	case <-em.CTX.Done():
		return em.CTX.Err()
	default:
		return nil
	}
}

func (em *EventManager) Shutdown() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for _, subs := range em.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	close(em.eventChan)
}

func (em *EventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	select {
	case <-em.CTX.Done():
		return em.CTX.Err()
	case em.eventChan <- interfaces.Event{Type: eventType, Data: data}:
		return nil
	default:
		return fmt.Errorf("failed to publish event: %v, channel full", eventType)
	}
}

func (em *EventManager) Subscribe(eventType interfaces.EventType) (<-chan interfaces.Event, error) {
	select {
	case <-em.CTX.Done():
		return nil, em.CTX.Err()
	default:
		ch := make(chan interfaces.Event, 1000)
		em.mu.Lock()
		defer em.mu.Unlock()
		em.subscribers[eventType] = append(em.subscribers[eventType], ch)
		return ch, nil
	}
}

func (em *EventManager) dispatch(event interfaces.Event) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	if subscribers, ok := em.subscribers[event.Type]; ok {
		for _, ch := range subscribers {
			select {
			case ch <- event:
				// success
			default:
				// channel is full, skip this subscriber
			}
		}
	}
}

func (em *EventManager) Unsubscribe(eventType interfaces.EventType, ch <-chan interfaces.Event) error {
	select {
	case <-em.CTX.Done():
		return em.CTX.Err()
	default:
		em.mu.Lock()
		defer em.mu.Unlock()
		if subscribers, ok := em.subscribers[eventType]; ok {
			for i, subscriber := range subscribers {
				if subscriber == ch {
					close(subscriber)
					em.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
					break
				}
			}
		}
		return nil
	}
}

var _ interfaces.EventManagerInterface = (*EventManager)(nil)
