package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ajkula/shmup/interfaces"
)

type MockEventManager struct {
	mu              sync.RWMutex
	subscribers     map[interfaces.EventType][]chan interfaces.Event
	publishedEvents []interfaces.Event
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewMockEventManager() *MockEventManager {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	return &MockEventManager{
		subscribers:     make(map[interfaces.EventType][]chan interfaces.Event),
		publishedEvents: []interfaces.Event{},
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (m *MockEventManager) Initialize(ctx context.Context) error {
	m.cancel()
	m.ctx, m.cancel = context.WithCancel(ctx)
	return nil
}

func (m *MockEventManager) Update(deltaTime float64) error {
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
		return nil
	}
}

func (m *MockEventManager) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockEventManager) Shutdown() {
	m.cancel()
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, subs := range m.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	m.publishedEvents = []interfaces.Event{}
}

func (m *MockEventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	event := interfaces.Event{Type: eventType, Data: data}
	m.publishedEvents = append(m.publishedEvents, event)
	subscribers, ok := m.subscribers[eventType]
	if !ok {
		return fmt.Errorf("failed to Publish %d", eventType)
	}
	for _, ch := range subscribers {
		select {
		case ch <- event:
			// success
		default:
			// channel is full, skip this subscriber
		}
	}
	return nil
}

func (m *MockEventManager) Subscribe(eventType interfaces.EventType) (<-chan interfaces.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch := make(chan interfaces.Event, 100)
	m.subscribers[eventType] = append(m.subscribers[eventType], ch)
	return ch, nil
}

func (m *MockEventManager) Unsubscribe(eventType interfaces.EventType, ch <-chan interfaces.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if subscribers, ok := m.subscribers[eventType]; ok {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				close(subscriber)
				m.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *MockEventManager) GetPublishedEvents() []interfaces.Event {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]interfaces.Event{}, m.publishedEvents...)
}

func (m *MockEventManager) ClearPublishedEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedEvents = []interfaces.Event{}
}

var _ interfaces.EventManagerInterface = (*MockEventManager)(nil)
