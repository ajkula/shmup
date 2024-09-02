package manager

import (
	"context"
	"testing"
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
)

func TestNewBulletManager(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)

	if bm == nil {
		t.Fatal("NewBulletManager returned nil")
	}
	if bm.eventManager != eventManager {
		t.Error("EventManager not set correctly")
	}
	if len(bm.bullets) != 0 {
		t.Error("Initial bullets slice should be empty")
	}
}

func TestBulletManagerInitialize(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize returned an error: %v", err)
	}
	if bm.CTX != ctx {
		t.Error("Context not set correctly")
	}
	if len(bm.eventChannels) != 2 {
		t.Errorf("Expected 2 event channels, got %d", len(bm.eventChannels))
	}
	if _, ok := bm.eventChannels[interfaces.BulletCreated]; !ok {
		t.Error("BulletCreated event channel not initialized")
	}
	if _, ok := bm.eventChannels[interfaces.BulletDestroyed]; !ok {
		t.Error("BulletDestroyed event channel not initialized")
	}
}

func TestBulletManagerUpdate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	bullet1 := mocks.NewMockBullet(100, 100, true, eventManager)
	bullet2 := mocks.NewMockBullet(200, 200, false, eventManager)

	bm.AddBullet(bullet1)
	bm.AddBullet(bullet2)

	done := make(chan bool)
	go func() {
		err := bm.Update(0.16)
		if err != nil {
			t.Errorf("Update returned an error: %v", err)
		}
		done <- true
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	case <-done:
		// test ok
	}

	if !bullet1.UpdateCalled || !bullet2.UpdateCalled {
		t.Error("Update not called on all bullets")
	}

	if len(bm.bullets) != 2 {
		t.Errorf("Expected 2 bullets, got %d", len(bm.bullets))
	}
}

func TestBulletManagerHandleBulletDestroyed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	// scenario 1: Bullet hors limites
	bullet1 := mocks.NewMockBullet(100, 100, true, eventManager)
	bm.AddBullet(bullet1)
	bullet1.SetOutOfBounds(true)

	// scenario 2: Bullet en collision
	bullet2 := mocks.NewMockBullet(200, 200, false, eventManager)
	bm.AddBullet(bullet2)
	enemy := &mocks.MockEnemy{}
	bullet2.OnCollision(enemy)

	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("First Update returned an error: %v", err)
	}

	// chan to check test finished
	done := make(chan bool)
	go func() {
		err = bm.Update(0.16)
		if err != nil {
			t.Errorf("Second Update returned an error: %v", err)
		}

		if len(bm.bullets) != 0 {
			t.Errorf("Expected 0 bullets after destruction, got %d", len(bm.bullets))
		}

		events := eventManager.GetPublishedEvents()
		destroyedEvents := 0
		for _, e := range events {
			if e.Type == interfaces.BulletDestroyed {
				destroyedEvents++
			}
		}
		if destroyedEvents != 2 {
			t.Errorf("Expected 2 BulletDestroyed events, got %d", destroyedEvents)
		}

		done <- true
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test timed out")
	case <-done:
		// test ok
	}

	bm.Shutdown()
	eventManager.Shutdown()
}

func TestBulletManagerHandleBulletCreated(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()
	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	if len(bm.bullets) != 0 {
		t.Errorf("Expected 0 bullets initially, got %d", len(bm.bullets))
	}

	bullet := mocks.NewMockBullet(100, 100, false, eventManager)
	eventManager.Publish(interfaces.BulletCreated, bullet)

	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if len(bm.bullets) != 1 {
		t.Errorf("Expected 1 bullet after BulletCreated event, got %d", len(bm.bullets))
	}

	// check instances == celle tirée
	if len(bm.bullets) > 0 && bm.bullets[0] != bullet {
		t.Error("The added bullet is not the one we created")
	}
}

func TestBulletManagerRemoveBullet(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)

	bullet := mocks.NewMockBullet(100, 100, true, eventManager)
	bm.AddBullet(bullet)
	bm.RemoveBullet(bullet)

	if len(bm.bullets) != 0 {
		t.Errorf("Expected 0 bullets after removal, got %d", len(bm.bullets))
	}
}

func TestBulletManagerShutdown(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)

	bullet := mocks.NewMockBullet(100, 100, false, eventManager)
	bm.AddBullet(bullet)
	bm.Shutdown()

	if len(bm.bullets) != 0 {
		t.Error("Shutdown did not clear bullets")
	}
}

func TestBulletManagerUpdateContextCancellation(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())
	bm.Initialize(ctx)

	cancel()

	err := bm.Update(0.16)
	if err == nil {
		t.Fatal("Update should have returned an error due to cancelled context")
	}
}

func TestBulletManagerConcurrency(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bm.Initialize(ctx)

	const numOperations = 1000
	done := make(chan bool)

	go func() {
		for i := 0; i < numOperations; i++ {
			bm.AddBullet(mocks.NewMockBullet(float64(i), float64(i), false, eventManager))
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < numOperations; i++ {
			err := bm.Update(0.16)
			if err != nil {
				t.Errorf("Update returned an error: %v", err)
			}
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	<-done
	<-done

	bm.mu.Lock()
	defer bm.mu.Unlock()
	if len(bm.bullets) != numOperations {
		t.Errorf("Expected %d bullets, got %d", numOperations, len(bm.bullets))
	}
}
