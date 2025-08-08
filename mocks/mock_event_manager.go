package mocks

import (
	"context"
	"sync"

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
	ctx, cancel := context.WithCancel(context.Background())
	return &MockEventManager{
		subscribers:     make(map[interfaces.EventType][]chan interfaces.Event),
		publishedEvents: make([]interfaces.Event, 0),
		ctx:             ctx,
		cancel:          cancel,
	}
}

func (m *MockEventManager) Initialize(ctx context.Context) error {
	if m.cancel != nil {
		m.cancel()
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	return nil
}

func (m *MockEventManager) Update(deltaTime float64) error {
	return nil
}

func (m *MockEventManager) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockEventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	event := interfaces.Event{Type: eventType, Data: data}
	m.publishedEvents = append(m.publishedEvents, event)

	if subscribers, exists := m.subscribers[eventType]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
				// Success
			default:
				// noop
			}
		}
	}

	return nil
}

func (m *MockEventManager) Subscribe(eventType interfaces.EventType) (<-chan interfaces.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.subscribers == nil {
		m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	}
	if m.ctx == nil {
		m.ctx = context.Background()
	}

	ch := make(chan interfaces.Event, 10000)
	m.subscribers[eventType] = append(m.subscribers[eventType], ch)
	return ch, nil
}

func (m *MockEventManager) Unsubscribe(eventType interfaces.EventType, ch <-chan interfaces.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if subscribers, ok := m.subscribers[eventType]; ok {
		for i, subscriber := range subscribers {
			if subscriber == ch {
				m.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)
				break
			}
		}
	}
	return nil
}

func (m *MockEventManager) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, subscribers := range m.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
	}
	m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	m.publishedEvents = make([]interfaces.Event, 0)
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
