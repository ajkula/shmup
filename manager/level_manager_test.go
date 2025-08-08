package manager

import (
	"context"
	"testing"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
)

func TestNewLevelManager(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)

	if lm == nil {
		t.Fatal("NewLevelManager returned nil")
	}
	if lm.currentLevel != 1 {
		t.Errorf("Initial level should be 1, got %d", lm.currentLevel)
	}
	if lm.difficulty != 1.0 {
		t.Errorf("Initial difficulty should be 1.0, got %f", lm.difficulty)
	}
}

func TestLevelManagerInitialize(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)
	ctx := context.Background()

	err := lm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize returned an error: %v", err)
	}
	if lm.CTX != ctx {
		t.Error("Context not set correctly")
	}
	if len(lm.eventChannels) != 1 {
		t.Errorf("Expected 1 event channel, got %d", len(lm.eventChannels))
	}
}

func TestLevelManagerUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := lm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize LevelManager: %v", err)
	}

	eventManager.Publish(interfaces.LevelEvent, 1)

	err = lm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if lm.GetLevel() != 2 {
		t.Errorf("Expected level to be 2, got %d", lm.GetLevel())
	}
}

func TestLevelManagerAdvanceLevel(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)

	lm.AdvanceLevel(2)

	if lm.GetLevel() != 3 {
		t.Errorf("Expected level to be 3, got %d", lm.GetLevel())
	}

	expectedDifficulty := 1.2
	if lm.GetDifficulty() != expectedDifficulty {
		t.Errorf("Expected difficulty to be %f, got %f", expectedDifficulty, lm.GetDifficulty())
	}

	events := eventManager.GetPublishedEvents()
	if len(events) != 1 || events[0].Type != interfaces.LevelChanged {
		t.Error("Expected LevelChanged to be published")
	}
}

func TestLevelManagerShutdown(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)
	ctx := context.Background()
	lm.Initialize(ctx)

	lm.AdvanceLevel(5)
	lm.Shutdown()

	if lm.GetLevel() != 1 {
		t.Errorf("Expected level to be reset to 1, got %d", lm.GetLevel())
	}

	if lm.GetDifficulty() != 1.0 {
		t.Errorf("Expected difficulty to be reset to 1.0, got %f", lm.GetDifficulty())
	}

	if lm.eventChannels != nil {
		t.Error("Event channels should be nil after shutdown")
	}
}

func TestLevelManagerConcurrency(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	lm := NewLevelManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := lm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize LevelManager: %v", err)
	}

	const numOperations = 1000
	const batchSize = 50

	for i := 0; i < numOperations; i++ {
		eventManager.Publish(interfaces.LevelEvent, 1)

		if i%batchSize == 0 {
			lm.Update(0.016)
		}
	}

	for i := 0; i < 100; i++ {
		lm.Update(0.016)
		if lm.GetLevel() >= numOperations+1 {
			break
		}
	}

	expectedLevel := numOperations + 1
	actualLevel := lm.GetLevel()

	if actualLevel != expectedLevel {
		t.Errorf("Expected level to be %d, got %d", expectedLevel, actualLevel)
	}

	lm.Shutdown()
}
