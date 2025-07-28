package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/ajkula/shmup/interfaces"
)

type MockEventManager struct {
	mu              sync.RWMutex
	subscribers     map[interfaces.EventType][]chan interfaces.Event
	publishedEvents []interfaces.Event
	eventQueue      chan interfaces.Event
	ctx             context.Context
	cancel          context.CancelFunc
	running         bool
}

func NewMockEventManager() *MockEventManager {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &MockEventManager{
		subscribers:     make(map[interfaces.EventType][]chan interfaces.Event),
		publishedEvents: make([]interfaces.Event, 0),
		eventQueue:      make(chan interfaces.Event, 1000),
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
	for {
		select {
		case event := <-m.eventQueue:
			m.dispatchEvent(event)
		case <-m.ctx.Done():
			return m.ctx.Err()
		default:
			return nil
		}
	}
}

func (m *MockEventManager) Run(ctx context.Context) error {
	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			m.mu.Lock()
			m.running = false
			m.mu.Unlock()
			return ctx.Err()
		default:
			m.processQueue()
			time.Sleep(time.Millisecond)
		}
	}
}

func (m *MockEventManager) processQueue() {
	for {
		select {
		case event := <-m.eventQueue:
			m.dispatchEvent(event)
		default:
			return
		}
	}
}

func (m *MockEventManager) dispatchEvent(event interfaces.Event) {
	m.mu.RLock()
	subscribers, exists := m.subscribers[event.Type]
	m.mu.RUnlock()

	if exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

func (m *MockEventManager) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}

	for {
		m.mu.RLock()
		running := m.running
		m.mu.RUnlock()
		if !running {
			break
		}
		time.Sleep(time.Millisecond)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, subs := range m.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}

	m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	m.running = false

	for {
		select {
		case <-m.eventQueue:
		default:
			return
		}
	}
}

func (m *MockEventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	event := interfaces.Event{Type: eventType, Data: data}

	m.mu.Lock()
	m.publishedEvents = append(m.publishedEvents, event)
	m.mu.Unlock()

	m.mu.RLock()
	isRunning := m.running
	m.mu.RUnlock()

	if isRunning {
		select {
		case m.eventQueue <- event:
		default:
			m.dispatchEvent(event)
		}
	} else {
		m.dispatchEvent(event)
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
