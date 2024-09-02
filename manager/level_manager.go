package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type LevelManager struct {
	core.BaseSystem
	currentLevel  int
	difficulty    float64
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
}

func NewLevelManager(eventManager interfaces.EventManagerInterface) *LevelManager {
	return &LevelManager{
		currentLevel:  1,
		difficulty:    1.0,
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (lm *LevelManager) Initialize(ctx context.Context) error {
	err := lm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.LevelEvent,
	}

	for _, eventType := range eventTypes {
		ch, err := lm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event type %v: %w", eventType, err)
		}
		lm.eventChannels[eventType] = ch
	}

	return nil
}

func (lm *LevelManager) Update(deltaTime float64) error {
	select {
	case <-lm.CTX.Done():
		return lm.CTX.Err()
	default:
		lm.processEvents()
		return nil
	}
}

func (lm *LevelManager) processEvents() {
	for eventType, ch := range lm.eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				lm.handleEvent(eventType, evt)
			default:
				// finished
				return
			}
		}
	}
}

func (lm *LevelManager) handleEvent(eventType interfaces.EventType, evt interfaces.Event) {
	switch eventType {
	case interfaces.LevelEvent:
		if levelChange, ok := evt.Data.(int); ok {
			lm.AdvanceLevel(levelChange)
		}
	}
}

func (lm *LevelManager) AdvanceLevel(levels int) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.currentLevel += levels
	lm.difficulty += float64(levels) * 0.1
	lm.eventManager.Publish(interfaces.LevelEvent, lm.currentLevel)
}

func (lm *LevelManager) GetLevel() int {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.currentLevel
}

func (lm *LevelManager) GetDifficulty() float64 {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	return lm.difficulty
}

func (lm *LevelManager) Shutdown() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	for eventType, ch := range lm.eventChannels {
		lm.eventManager.Unsubscribe(eventType, ch)
	}
	lm.eventChannels = nil
	lm.currentLevel = 1
	lm.difficulty = 1.0
}

var _ core.System = (*LevelManager)(nil)
