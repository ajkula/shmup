package manager

import (
	"context"
	"fmt"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type ScoreManager struct {
	core.BaseSystem
	score         int
	highScore     int
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
}

func NewScoreManager(eventManager interfaces.EventManagerInterface) *ScoreManager {
	return &ScoreManager{
		score:         0,
		highScore:     0,
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (sm *ScoreManager) Initialize(ctx context.Context) error {
	err := sm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	sm.eventChannels[interfaces.ScoreEvent], err = sm.eventManager.Subscribe(interfaces.ScoreEvent)
	if err != nil {
		return fmt.Errorf("failed to subscribe to ScoreEvent: %w", err)
	}

	return nil
}

func (sm *ScoreManager) Update(deltaTime float64) error {
	select {
	case <-sm.CTX.Done():
		return sm.CTX.Err()
	default:
		sm.processEvents()
		return nil
	}
}

func (sm *ScoreManager) processEvents() {
	for eventType, ch := range sm.eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				sm.handleEvent(eventType, evt)
			default:
				// no more events
				return
			}
		}
	}
}

func (sm *ScoreManager) handleEvent(eventType interfaces.EventType, evt interfaces.Event) {
	switch eventType {
	case interfaces.ScoreEvent:
		if scoreChange, ok := evt.Data.(int); ok {
			sm.AddScore(scoreChange)
		}
	}
}

func (sm *ScoreManager) AddScore(points int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.score += points
	if sm.score > sm.highScore {
		sm.highScore = sm.score
	}
}

func (sm *ScoreManager) GetScore() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.score
}

func (sm *ScoreManager) GetHighScore() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.highScore
}

func (sm *ScoreManager) ResetScore() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.score = 0
	sm.eventManager.Publish(interfaces.ScoreEvent, sm.score)
	fmt.Println("Score reset")
}

func (sm *ScoreManager) Shutdown() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for eventType, ch := range sm.eventChannels {
		sm.eventManager.Unsubscribe(eventType, ch)
	}
	sm.eventChannels = nil
	sm.score = 0
	sm.highScore = 0
	fmt.Println("ScoreManager shut down")
}

var _ core.System = (*ScoreManager)(nil)
