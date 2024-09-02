package manager

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
)

func TestNewScoreManager(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)

	if sm == nil {
		t.Fatal("NewScoreManager returned nil")
	}
	if sm.eventManager != eventManager {
		t.Error("EventManager not set correctly")
	}
	if sm.score != 0 {
		t.Error("Initial score should be 0")
	}
	if sm.highScore != 0 {
		t.Error("Initial high score should be 0")
	}
}

func TestScoreManagerInitialize(t *testing.T) {
	sm := NewScoreManager(mocks.NewMockEventManager())
	ctx := context.Background()

	err := sm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize returned an error: %v", err)
	}
	if sm.CTX != ctx {
		t.Error("Context not set correctly")
	}
}

func TestScoreManagerUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)
	ctx := context.Background()
	sm.Initialize(ctx)

	eventManager.Publish(interfaces.ScoreEvent, 100)

	err := sm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	if sm.GetScore() != 100 {
		t.Errorf("Expected score to be 100, got %d", sm.GetScore())
	}
}

func TestScoreManagerAddScore(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)
	ctx := context.Background()
	sm.Initialize(ctx)

	sm.AddScore(50)
	if sm.GetScore() != 50 {
		t.Errorf("Expected score to be 50, got %d", sm.GetScore())
	}

	sm.AddScore(30)
	if sm.GetScore() != 80 {
		t.Errorf("Expected score to be 80, got %d", sm.GetScore())
	}

	if sm.GetHighScore() != 80 {
		t.Errorf("Expected high score to be 80, got %d", sm.GetHighScore())
	}
}

func TestScoreManagerResetScore(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)
	ctx := context.Background()
	sm.Initialize(ctx)

	sm.AddScore(100)
	sm.ResetScore()

	if sm.GetScore() != 0 {
		t.Errorf("Expected score to be 0 after reset, got %d", sm.GetScore())
	}

	if sm.GetHighScore() != 100 {
		t.Errorf("Expected high score to remain 100 after reset, got %d", sm.GetHighScore())
	}
}

func TestScoreManagerConcurrency(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := sm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize ScoreManager: %v", err)
	}

	const numOperations = 1000
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			eventManager.Publish(interfaces.ScoreEvent, 1)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < numOperations; i++ {
			sm.Update(0.16)
		}
	}()

	wg.Wait()

	for i := 0; i < 10; i++ {
		sm.Update(0.16)
	}

	expectedScore := numOperations
	actualScore := sm.GetScore()
	t.Logf("Final score: %d", actualScore)
	if actualScore != expectedScore {
		t.Errorf("Expected score to be %d, got %d", expectedScore, actualScore)
	}

	sm.Shutdown()
}

func TestScoreManagerShutdown(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	sm := NewScoreManager(eventManager)
	ctx := context.Background()
	sm.Initialize(ctx)

	sm.AddScore(100)
	sm.Shutdown()

	if sm.GetScore() != 0 {
		t.Errorf("Expected score to be 0 after shutdown, got %d", sm.GetScore())
	}

	if sm.GetHighScore() != 0 {
		t.Errorf("Expected high score to be 0 after shutdown, got %d", sm.GetHighScore())
	}
}
