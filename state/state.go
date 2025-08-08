package state

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type GameState int

const (
	StateMainMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

type StateManager struct {
	core.BaseSystem
	currentState    int32
	eventManager    interfaces.EventManagerInterface
	eventChannels   map[interfaces.EventType]<-chan interfaces.Event
	mu              sync.RWMutex
	isShutdown      int32
}

func NewStateManager(eventManager interfaces.EventManagerInterface) *StateManager {
	return &StateManager{
		currentState:  int32(StateMainMenu),
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (sm *StateManager) Initialize(ctx context.Context) error {
	err := sm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.SystemTick,
		interfaces.GameStateChangeEvent,
	}

	for _, eventType := range eventTypes {
		ch, err := sm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event type %v: %w", eventType, err)
		}
		sm.eventChannels[eventType] = ch
	}

	go sm.eventListener()

	return nil
}

func (sm *StateManager) Update(deltaTime float64) error {
	select {
	case <-sm.CTX.Done():
		return sm.CTX.Err()
	default:
		return nil
	}
}

func (sm *StateManager) eventListener() {
	systemTickCh := sm.eventChannels[interfaces.SystemTick]
	stateChangeCh := sm.eventChannels[interfaces.GameStateChangeEvent]

	for {
		select {
		case <-sm.CTX.Done():
			return
		case _, ok := <-systemTickCh:
			if !ok {
				return
			}
			sm.processAllAvailableEvents()
			sm.updateCurrentState()
		case evt, ok := <-stateChangeCh:
			if !ok {
				return
			}
			sm.handleStateChangeEvent(evt)
		}
	}
}

func (sm *StateManager) processAllAvailableEvents() {
	if atomic.LoadInt32(&sm.isShutdown) == 1 {
		return
	}

	stateChangeCh := sm.eventChannels[interfaces.GameStateChangeEvent]
	for {
		select {
		case evt, ok := <-stateChangeCh:
			if !ok {
				return
			}
			sm.handleStateChangeEvent(evt)
		default:
			return
		}
	}
}

func (sm *StateManager) updateCurrentState() {
	currentState := GameState(atomic.LoadInt32(&sm.currentState))

	switch currentState {
	case StatePlaying:
		// Game playing logic
	case StatePaused:
		// Pause logic
	case StateMainMenu:
		// Main menu logic
	case StateGameOver:
		// Game over logic
	}
}

func (sm *StateManager) handleStateChangeEvent(evt interfaces.Event) {
	if newState, ok := evt.Data.(GameState); ok {
		sm.setState(newState)
	}
}

func (sm *StateManager) setState(state GameState) {
	currentState := GameState(atomic.LoadInt32(&sm.currentState))
	if currentState == state {
		return
	}

	sm.exitState(currentState)
	atomic.StoreInt32(&sm.currentState, int32(state))
	sm.enterState(state)

	sm.eventManager.Publish(interfaces.GameStateChangeEvent, state)
}

func (sm *StateManager) GetState() GameState {
	return GameState(atomic.LoadInt32(&sm.currentState))
}

func (sm *StateManager) exitState(state GameState) {
	// State-specific exit logic
}

func (sm *StateManager) enterState(state GameState) {
	// State-specific enter logic
}

func (sm *StateManager) RequestStateChange(state GameState) {
	sm.eventManager.Publish(interfaces.GameStateChangeEvent, state)
}

func (sm *StateManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&sm.isShutdown, 0, 1) {
		return
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	for eventType, ch := range sm.eventChannels {
		sm.eventManager.Unsubscribe(eventType, ch)
	}
	sm.eventChannels = nil

	atomic.StoreInt32(&sm.currentState, int32(StateMainMenu))
}

var _ core.System = (*StateManager)(nil)