package state

import (
	"context"
	"fmt"

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
	currentState    GameState
	eventManager    interfaces.EventManagerInterface
	stateChangeChan chan GameState
}

func NewStateManager(eventManager interfaces.EventManagerInterface) *StateManager {
	return &StateManager{
		currentState:    StateMainMenu,
		eventManager:    eventManager,
		stateChangeChan: make(chan GameState, 10),
	}
}

func (sm *StateManager) Initialize(ctx context.Context) error {
	sm.CTX = ctx
	return nil
}

func (sm *StateManager) Run(ctx context.Context) error {
	sm.CTX = ctx
	stateChan, err := sm.eventManager.Subscribe(interfaces.GameStateChangeEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to GameStateChangeEvent: %w", err)
	}
	defer sm.eventManager.Unsubscribe(interfaces.GameStateChangeEvent, stateChan)

	for {
		select {
		case <-sm.CTX.Done():
			return sm.CTX.Err()
		case newState := <-sm.stateChangeChan:
			sm.setState(newState)
		case evt := <-stateChan:
			if newState, ok := evt.Data.(GameState); ok {
				sm.stateChangeChan <- newState
			}
		}
	}
}

func (sm *StateManager) Update(deltaTime float64) error {
	select {
	case <-sm.CTX.Done():
		return sm.CTX.Err()
	default:
		// logique maj specifique etat actuel
		switch sm.currentState {
		case StatePlaying:
			// logique maj etat de jeu
		case StatePaused:
			// logique maj etat en pause
		}
	}
	return nil
}

func (sm *StateManager) Shutdown() {
	// cleanup
	close(sm.stateChangeChan)
}

func (sm *StateManager) setState(state GameState) {
	if sm.currentState == state {
		return
	}

	sm.exitState(sm.currentState)
	sm.currentState = state
	sm.enterState(state)

	sm.eventManager.Publish(interfaces.GameStateChangeEvent, state)
}

func (sm *StateManager) GetState() GameState {
	return sm.currentState
}

func (sm *StateManager) exitState(state GameState) {
	// logique de sortie specifique à chaque etat
}

func (sm *StateManager) enterState(state GameState) {
	// logique entrée spécifique à chaque etat
}

func (sm *StateManager) RequestStateChange(state GameState) {
	sm.stateChangeChan <- state
}
