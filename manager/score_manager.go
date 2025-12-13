package manager

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

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
	isShutdown    int32
}

func NewScoreManager(eventManager interfaces.EventManagerInterface) *ScoreManager {
	sm := &ScoreManager{
		score:         0,
		highScore:     0,
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}

	// Load high score from file
	sm.loadHighScore()

	return sm
}

func (sm *ScoreManager) Initialize(ctx context.Context) error {
	err := sm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.ScoreEvent,
	}

	for _, eventType := range eventTypes {
		ch, err := sm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event type %v: %w", eventType, err)
		}
		sm.eventChannels[eventType] = ch
	}

	return nil
}

func (sm *ScoreManager) Update(deltaTime float64) error {
	select {
	case <-sm.CTX.Done():
		return sm.CTX.Err()
	default:
		sm.processAllEvents()
		return nil
	}
}

func (sm *ScoreManager) processAllEvents() {
	if atomic.LoadInt32(&sm.isShutdown) == 1 {
		return
	}

	scoreEventCh := sm.eventChannels[interfaces.ScoreEvent]
	processed := 0
	maxProcess := 1000

	for processed < maxProcess {
		select {
		case evt, ok := <-scoreEventCh:
			if !ok {
				return
			}
			if scoreChange, ok := evt.Data.(int); ok {
				sm.AddScore(scoreChange)
			}
			processed++
		default:
			return
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
			// New high score! Save it
			sm.saveHighScore()
			fmt.Printf("🏆 NEW HIGH SCORE: %d\n", newScore)
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

func (sm *ScoreManager) loadHighScore() {
	data, err := os.ReadFile("highscore.dat")
	if err != nil {
		// File doesn't exist or can't be read, start with 0
		fmt.Println("No high score file found, starting fresh")
		return
	}

	scoreStr := strings.TrimSpace(string(data))
	highScore, err := strconv.ParseInt(scoreStr, 10, 64)
	if err != nil {
		fmt.Printf("Error parsing high score: %v\n", err)
		return
	}

	atomic.StoreInt64(&sm.highScore, highScore)
	fmt.Printf("Loaded high score: %d\n", highScore)
}

func (sm *ScoreManager) saveHighScore() {
	highScore := atomic.LoadInt64(&sm.highScore)
	err := os.WriteFile("highscore.dat", []byte(fmt.Sprintf("%d", highScore)), 0644)
	if err != nil {
		fmt.Printf("Error saving high score: %v\n", err)
		return
	}
}

func (sm *ScoreManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&sm.isShutdown, 0, 1) {
		return
	}

	sm.mu.Lock()
	for eventType, ch := range sm.eventChannels {
		sm.eventManager.Unsubscribe(eventType, ch)
	}
	sm.eventChannels = nil
	sm.mu.Unlock()

	// Save high score before shutdown
	sm.saveHighScore()

	atomic.StoreInt64(&sm.score, 0)

	fmt.Println("ScoreManager shut down")
}
