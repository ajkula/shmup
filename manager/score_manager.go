package manager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type ScoreManager struct {
	core.BaseSystem
	score         int64
	highScore     int64
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
	isShutdown    int32
}

func NewScoreManager(eventManager interfaces.EventManagerInterface) *ScoreManager {
	return &ScoreManager{
		score:         0,
		highScore:     0,
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
		shutdownCh:    make(chan struct{}),
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

	sm.wg.Add(1)
	go sm.eventProcessor()

	return nil
}

func (sm *ScoreManager) Update(deltaTime float64) error {
	select {
	case <-sm.CTX.Done():
		return sm.CTX.Err()
	default:
		sm.processAllAvailableEvents()
		return nil
	}
}

func (sm *ScoreManager) eventProcessor() {
	defer sm.wg.Done()

	for {
		select {
		case <-sm.CTX.Done():
			return
		case <-sm.shutdownCh:
			return
		default:
			sm.processAllAvailableEvents()
			time.Sleep(16 * time.Millisecond) // ~60fps, prevent CPU burning
		}
	}
}

func (sm *ScoreManager) processAllAvailableEvents() {
	sm.mu.RLock()
	channels := make(map[interfaces.EventType]<-chan interfaces.Event)
	for k, v := range sm.eventChannels {
		channels[k] = v
	}
	sm.mu.RUnlock()

	for eventType, ch := range channels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					sm.mu.Lock()
					delete(sm.eventChannels, eventType)
					sm.mu.Unlock()
					goto nextChannel
				}
				sm.handleEvent(eventType, evt)
			default:
				goto nextChannel
			}
		}
	nextChannel:
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
	newScore := atomic.AddInt64(&sm.score, int64(points))

	for {
		currentHigh := atomic.LoadInt64(&sm.highScore)
		if newScore <= currentHigh {
			break
		}
		if atomic.CompareAndSwapInt64(&sm.highScore, currentHigh, newScore) {
			break
		}
	}
}

func (sm *ScoreManager) GetScore() int {
	return int(atomic.LoadInt64(&sm.score))
}

func (sm *ScoreManager) GetHighScore() int {
	return int(atomic.LoadInt64(&sm.highScore))
}

func (sm *ScoreManager) ResetScore() {
	atomic.StoreInt64(&sm.score, 0)
	sm.eventManager.Publish(interfaces.ScoreEvent, 0)
	fmt.Println("Score reset")
}

func (sm *ScoreManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&sm.isShutdown, 0, 1) {
		return
	}

	close(sm.shutdownCh)
	sm.wg.Wait()

	sm.mu.Lock()
	for eventType, ch := range sm.eventChannels {
		sm.eventManager.Unsubscribe(eventType, ch)
	}
	sm.eventChannels = nil
	sm.mu.Unlock()

	atomic.StoreInt64(&sm.score, 0)
	atomic.StoreInt64(&sm.highScore, 0)

	fmt.Println("ScoreManager shut down")
}

var _ core.System = (*ScoreManager)(nil)
