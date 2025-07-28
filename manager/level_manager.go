package manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type LevelManager struct {
	core.BaseSystem
	currentLevel  int64
	difficulty    float64
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
	isShutdown    int32
}

func NewLevelManager(eventManager interfaces.EventManagerInterface) *LevelManager {
	return &LevelManager{
		currentLevel:  1,
		difficulty:    1.0,
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
		shutdownCh:    make(chan struct{}),
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

	lm.wg.Add(1)
	go lm.eventProcessor()

	return nil
}

func (lm *LevelManager) Update(deltaTime float64) error {
	select {
	case <-lm.CTX.Done():
		return lm.CTX.Err()
	default:
		// noop
		return nil
	}
}

func (lm *LevelManager) eventProcessor() {
	defer lm.wg.Done()

	for {
		select {
		case <-lm.CTX.Done():
			return
		case <-lm.shutdownCh:
			return
		default:
			lm.processAllAvailableEvents()
		}
	}
}

func (lm *LevelManager) processAllAvailableEvents() {
	for eventType, ch := range lm.eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					lm.mu.Lock()
					delete(lm.eventChannels, eventType)
					lm.mu.Unlock()
					return
				}
				lm.handleEvent(eventType, evt)
			default:
				goto nextChannel
			}
		}
	nextChannel:
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
	newLevel := atomic.AddInt64(&lm.currentLevel, int64(levels))
	lm.difficulty += float64(levels) * 0.1
	lm.mu.Unlock()

	err := lm.eventManager.Publish(interfaces.LevelChanged, int(newLevel))
	if err != nil {
		fmt.Printf("Failed to publish LevelChanged: %v\n", err)
	}
}

func (lm *LevelManager) GetLevel() int {
	return int(atomic.LoadInt64(&lm.currentLevel))
}

func (lm *LevelManager) GetDifficulty() float64 {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.difficulty
}

func (lm *LevelManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&lm.isShutdown, 0, 1) {
		return
	}

	close(lm.shutdownCh)
	lm.wg.Wait()

	lm.mu.Lock()
	defer lm.mu.Unlock()
	for eventType, ch := range lm.eventChannels {
		lm.eventManager.Unsubscribe(eventType, ch)
	}
	lm.eventChannels = nil

	atomic.StoreInt64(&lm.currentLevel, 1)
	lm.difficulty = 1.0
}

var _ core.System = (*LevelManager)(nil)
