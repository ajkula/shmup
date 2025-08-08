package manager

import (
	"context"
	"errors"
	"testing"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
)

func TestNewEnemyManager(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)

	if em == nil {
		t.Fatal("NewEnemyManager returned nil")
	}
	if em.eventManager != eventManager {
		t.Error("EventManager not set correctly")
	}
	if em.GetEnemyCount() != 0 {
		t.Error("Initial enemies slice should be empty")
	}
}

func TestEnemyManagerInitialize(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	ctx := context.Background()

	err := em.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize returned an error: %v", err)
	}
	if em.CTX != ctx {
		t.Error("Context not set correctly")
	}
}

func TestEnemyManagerUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)
	ctx := context.Background()
	em.Initialize(ctx)

	// Create mock enemies via events
	enemy1 := &mocks.MockEnemy{Alive: true}
	enemy2 := &mocks.MockEnemy{Alive: false}

	// Publish enemy creation events
	eventManager.Publish(interfaces.EnemyCreated, enemy1)
	eventManager.Publish(interfaces.EnemyCreated, enemy2)

	// Process events dans Update
	err := em.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if !enemy1.UpdateCalled {
		t.Error("Update not called on alive enemy")
	}

	// After update, dead enemy should be removed
	finalCount := em.GetEnemyCount()
	if finalCount != 1 {
		t.Errorf("Expected 1 alive enemy after update, got %d", finalCount)
	}
}

func TestEnemyManagerUpdateError(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)
	ctx := context.Background()
	em.Initialize(ctx)

	enemy := &mocks.MockEnemy{Alive: true, UpdateError: errors.New("update error")}
	eventManager.Publish(interfaces.EnemyCreated, enemy)

	em.Update(0.16)

	// Should not return an error since enemy updates are handled gracefully
	err := em.Update(0.16)
	if err != nil {
		t.Fatalf("Update should not return error for enemy update failures: %v", err)
	}
}

func TestEnemyManagerEnemyEvents(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)
	ctx := context.Background()
	em.Initialize(ctx)

	enemy := &mocks.MockEnemy{Alive: true}

	// Test enemy creation
	eventManager.Publish(interfaces.EnemyCreated, enemy)
	em.Update(0.16) // Process events

	if em.GetEnemyCount() != 1 {
		t.Error("Enemy not added via event")
	}

	// Test enemy destruction
	eventManager.Publish(interfaces.EnemyDestroyed, enemy)
	em.Update(0.16) // Process events

	if em.GetEnemyCount() != 0 {
		t.Error("Enemy not removed via event")
	}
}

func TestEnemyManagerGetRenderableEntities(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)
	ctx := context.Background()
	em.Initialize(ctx)

	enemy1 := &mocks.MockEnemy{Alive: true}
	enemy2 := &mocks.MockEnemy{Alive: false}

	eventManager.Publish(interfaces.EnemyCreated, enemy1)
	eventManager.Publish(interfaces.EnemyCreated, enemy2)

	em.Update(0.16) // Process events

	renderables := em.GetRenderableEntities()

	// Only alive enemies should be rendered
	if len(renderables) != 1 {
		t.Errorf("Expected 1 renderable enemy, got %d", len(renderables))
	}
}

func TestEnemyManagerShutdown(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	em := NewEnemyManager(eventManager)
	ctx := context.Background()

	err := em.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EnemyManager: %v", err)
	}

	enemy := &mocks.MockEnemy{Alive: true}
	eventManager.Publish(interfaces.EnemyCreated, enemy)

	em.Update(0.016)

	if em.GetEnemyCount() != 1 {
		t.Errorf("Expected 1 enemy before shutdown, got %d", em.GetEnemyCount())
	}

	em.Shutdown()

	if em.GetEnemyCount() != 0 {
		t.Error("Shutdown did not clear enemies")
	}
}

func TestEnemyManagerUpdateContextCancellation(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	ctx, cancel := context.WithCancel(context.Background())
	em.Initialize(ctx)

	cancel()

	err := em.Update(0.16)
	if err == nil {
		t.Fatal("Update should have returned an error due to cancelled context")
	}
}
