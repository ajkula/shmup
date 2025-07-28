package event

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type EventManager struct {
	core.BaseSystem
	eventChan   chan interfaces.Event
	subscribers map[interfaces.EventType][]*subscriberNode
	mu          sync.RWMutex
	shutdownCh  chan struct{}
	wg          sync.WaitGroup
	isShutdown  int32 // atomic
}

type subscriberNode struct {
	ch     chan interfaces.Event
	active int32
}

func NewEventManager() interfaces.EventManagerInterface {
	return &EventManager{
		eventChan:   make(chan interfaces.Event, 5000),
		subscribers: make(map[interfaces.EventType][]*subscriberNode),
		shutdownCh:  make(chan struct{}),
	}
}

func (em *EventManager) Initialize(ctx context.Context) error {
	if err := em.BaseSystem.Initialize(ctx); err != nil {
		return err
	}

	em.wg.Add(1)
	go em.processEvents()
	return nil
}

func (em *EventManager) Run(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (em *EventManager) processEvents() {
	defer em.wg.Done()

	for {
		select {
		case <-em.CTX.Done():
			return
		case <-em.shutdownCh:
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
	if !atomic.CompareAndSwapInt32(&em.isShutdown, 0, 1) {
		return
	}

	close(em.shutdownCh)
	em.wg.Wait()

	em.mu.Lock()
	defer em.mu.Unlock()

	for _, nodes := range em.subscribers {
		for _, node := range nodes {
			atomic.StoreInt32(&node.active, 0)
			close(node.ch)
		}
	}

	close(em.eventChan)
	for range em.eventChan {
		// drain remaining events
	}
}

func (em *EventManager) Publish(eventType interfaces.EventType, data interface{}) error {
	if atomic.LoadInt32(&em.isShutdown) == 1 {
		return fmt.Errorf("event manager is shutdown")
	}

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
	if atomic.LoadInt32(&em.isShutdown) == 1 {
		return nil, fmt.Errorf("event manager is shutdown")
	}

	select {
	case <-em.CTX.Done():
		return nil, em.CTX.Err()
	default:
		ch := make(chan interfaces.Event, 1000)
		node := &subscriberNode{
			ch:     ch,
			active: 1,
		}

		em.mu.Lock()
		em.subscribers[eventType] = append(em.subscribers[eventType], node)
		em.mu.Unlock()

		return ch, nil
	}
}

func (em *EventManager) dispatch(event interfaces.Event) {
	em.mu.RLock()
	nodes := em.subscribers[event.Type]

	nodesCopy := make([]*subscriberNode, len(nodes))
	copy(nodesCopy, nodes)
	em.mu.RUnlock()

	for _, node := range nodesCopy {
		if atomic.LoadInt32(&node.active) == 1 {
			select {
			case node.ch <- event:
				// success
			default:
				// noop
			}
		}
	}
}

func (em *EventManager) Unsubscribe(eventType interfaces.EventType, ch <-chan interfaces.Event) error {
	if atomic.LoadInt32(&em.isShutdown) == 1 {
		return nil // ignore lors du shutdown
	}

	select {
	case <-em.CTX.Done():
		return em.CTX.Err()
	default:
		em.mu.Lock()
		defer em.mu.Unlock()

		if nodes, ok := em.subscribers[eventType]; ok {
			for i, node := range nodes {
				if node.ch == ch {

					atomic.StoreInt32(&node.active, 0)
					close(node.ch)

					em.subscribers[eventType] = append(nodes[:i], nodes[i+1:]...)
					break
				}
			}
		}
		return nil
	}
}

var _ interfaces.EventManagerInterface = (*EventManager)(nil)
