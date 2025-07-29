package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/ajkula/shmup/interfaces"
)

// MockEventManager - SIMPLE et SYNCHRONE pour tests de logique métier
type MockEventManager struct {
	mu              sync.RWMutex
	subscribers     map[interfaces.EventType][]chan interfaces.Event
	publishedEvents []interfaces.Event // Métriques pour assertions
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

// Update - Traitement synchrone et déterministe
func (m *MockEventManager) Update(deltaTime float64) error {
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
		// Traitement synchrone : pas de goroutines, pas de timing
		return nil
	}
}

// Run - Simple passthrough pour compatibilité, mais pas utilisé dans les tests
func (m *MockEventManager) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Publish - SYNCHRONE : dispatch immédiat
func (m *MockEventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Auto-initialize if not properly constructed
	if m.subscribers == nil {
		m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	}
	if m.publishedEvents == nil {
		m.publishedEvents = make([]interfaces.Event, 0)
	}
	if m.ctx == nil {
		m.ctx = context.Background()
	}

	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
		event := interfaces.Event{Type: eventType, Data: data}
		m.publishedEvents = append(m.publishedEvents, event)
		m.dispatchEvent(event) // Already under lock
		return nil
	}
}

func (m *MockEventManager) dispatchEvent(event interfaces.Event) {
	// Called from Publish which already holds the lock
	subscribers := m.subscribers[event.Type]
	subscribersCopy := make([]chan interfaces.Event, len(subscribers))
	copy(subscribersCopy, subscribers)

	// Dispatch synchrone - pas de select/default qui peut dropper
	for _, ch := range subscribersCopy {
		select {
		case ch <- event:
			// Success
		case <-m.ctx.Done():
			return
		default:
			// shouldn't happen during tests
			fmt.Println("CHANNEL FULL IN DISPATCH EVENT: SHOULDN'T HAPPEN!")
		}
	}
}

func (m *MockEventManager) Subscribe(eventType interfaces.EventType) (<-chan interfaces.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Auto-initialize if not properly constructed
	if m.subscribers == nil {
		m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	}
	if m.publishedEvents == nil {
		m.publishedEvents = make([]interfaces.Event, 0)
	}
	if m.ctx == nil {
		m.ctx = context.Background()
	}

	select {
	case <-m.ctx.Done():
		return nil, m.ctx.Err()
	default:
		ch := make(chan interfaces.Event, 1000)
		m.subscribers[eventType] = append(m.subscribers[eventType], ch)
		return ch, nil
	}
}

func (m *MockEventManager) Unsubscribe(eventType interfaces.EventType, ch <-chan interfaces.Event) error {
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	default:
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
}

// Shutdown - Nettoyage propre
func (m *MockEventManager) Shutdown() {
	if m.cancel != nil {
		m.cancel()
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Fermer tous les channels
	for _, subscribers := range m.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
	}

	// Reset pour prochain test
	m.subscribers = make(map[interfaces.EventType][]chan interfaces.Event)
	m.publishedEvents = make([]interfaces.Event, 0)
}

// Métriques pour assertions de test
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
