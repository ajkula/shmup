package manager

import (
	"context"
	"errors"
	"testing"
	"time"

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
	if len(em.enemies) != 0 {
		t.Error("Initial enemies slice should be empty")
	}
	if len(em.formations) != 0 {
		t.Error("Initial formations slice should be empty")
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
	em := NewEnemyManager(mocks.NewMockEventManager())
	ctx := context.Background()
	em.Initialize(ctx)

	enemy1 := &mocks.MockEnemy{Alive: true}
	enemy2 := &mocks.MockEnemy{Alive: false}
	formation := &mocks.MockFormation{}

	em.AddEnemy(enemy1)
	em.AddEnemy(enemy2)
	em.AddFormation(formation)

	err := em.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if !enemy1.UpdateCalled {
		t.Error("Update not called on alive enemy")
	}
	if !formation.UpdateCalled {
		t.Error("Update not called on formation")
	}

	if len(em.enemies) != 1 {
		t.Error("Dead enemy not removed")
	}
}

func TestEnemyManagerUpdateError(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	ctx := context.Background()
	em.Initialize(ctx)

	enemy := &mocks.MockEnemy{Alive: true, UpdateError: errors.New("update error")}
	em.AddEnemy(enemy)

	err := em.Update(0.16)
	if err == nil {
		t.Fatal("Update should have returned an error")
	}
}

func TestEnemyManagerAddRemoveEnemy(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	enemy := &mocks.MockEnemy{}

	em.AddEnemy(enemy)
	if len(em.enemies) != 1 {
		t.Error("Enemy not added")
	}

	em.RemoveEnemy(enemy)
	if len(em.enemies) != 0 {
		t.Error("Enemy not removed")
	}
}

func TestEnemyManagerAddRemoveFormation(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	formation := &mocks.MockFormation{}

	em.AddFormation(formation)
	if len(em.formations) != 1 {
		t.Error("Formation not added")
	}

	em.RemoveFormation(formation)
	if len(em.formations) != 0 {
		t.Error("Formation not removed")
	}
}

func TestEnemyManagerShutdown(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	enemy := &mocks.MockEnemy{}
	formation := &mocks.MockFormation{}

	em.AddEnemy(enemy)
	em.AddFormation(formation)

	em.Shutdown()

	if len(em.enemies) != 0 || len(em.formations) != 0 {
		t.Error("Shutdown did not clear enemies and formations")
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

func TestEnemyManagerConcurrency(t *testing.T) {
	em := NewEnemyManager(mocks.NewMockEventManager())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	em.Initialize(ctx)

	const numOperations = 1000
	done := make(chan bool)

	go func() {
		for i := 0; i < numOperations; i++ {
			em.AddEnemy(&mocks.MockEnemy{Alive: i%2 == 0})
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < numOperations; i++ {
			err := em.Update(0.16)
			if err != nil {
				t.Errorf("Update returned an error: %v", err)
			}
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	<-done
	<-done

	em.mu.Lock()
	defer em.mu.Unlock()
	if len(em.enemies) > numOperations/2 {
		t.Errorf("Expected at most %d enemies, got %d", numOperations/2, len(em.enemies))
	}

	em.Shutdown()
}
