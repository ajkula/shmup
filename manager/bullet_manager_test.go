package manager

import (
	"context"
	"testing"

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
	if bm.GetBulletCount() != 0 {
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

	// Check that we have the expected event channels
	expectedChannels := 6 // SystemTick, BulletCreated, BulletDestroyed, PlayerShot, EnemyShot, BossShot
	if len(bm.eventChannels) != expectedChannels {
		t.Errorf("Expected %d event channels, got %d", expectedChannels, len(bm.eventChannels))
	}
}

func TestBulletManagerShootEvents(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	// Create a mock player to shoot
	player := &mocks.MockEnemy{} // MockEnemy implements GameEntity

	// Publish player shot event
	eventManager.Publish(interfaces.PlayerShot, player)

	// Process the event synchronously
	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	// Check that a bullet was created
	if bm.GetBulletCount() != 1 {
		t.Errorf("Expected 1 bullet after PlayerShot event, got %d", bm.GetBulletCount())
	}

	// Check that bullet creation event was published
	events := eventManager.GetPublishedEvents()
	bulletCreatedCount := 0
	for _, e := range events {
		if e.Type == interfaces.BulletCreated {
			bulletCreatedCount++
		}
	}
	if bulletCreatedCount < 1 {
		t.Errorf("Expected at least 1 BulletCreated event, got %d", bulletCreatedCount)
	}
}

func TestBulletManagerUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	bullet1 := mocks.NewMockBullet(100, 100, true, eventManager)
	bullet2 := mocks.NewMockBullet(200, 200, false, eventManager)

	// Add bullets via events
	eventManager.Publish(interfaces.BulletCreated, bullet1)
	eventManager.Publish(interfaces.BulletCreated, bullet2)

	// Process events synchronously
	err = bm.Update(0.16)
	if err != nil {
		t.Errorf("Update returned an error: %v", err)
	}

	if !bullet1.UpdateCalled || !bullet2.UpdateCalled {
		t.Error("Update not called on all bullets")
	}

	if bm.GetBulletCount() != 2 {
		t.Errorf("Expected 2 bullets, got %d", bm.GetBulletCount())
	}
}

func TestBulletManagerHandleBulletDestroyed(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	// Clear any existing events
	eventManager.ClearPublishedEvents()

	// Create bullets
	bullet1 := mocks.NewMockBullet(100, 100, true, eventManager)
	bullet2 := mocks.NewMockBullet(200, 200, false, eventManager)

	// Add bullets
	eventManager.Publish(interfaces.BulletCreated, bullet1)
	eventManager.Publish(interfaces.BulletCreated, bullet2)
	bm.Update(0.16) // Process creation events

	if bm.GetBulletCount() != 2 {
		t.Fatalf("Expected 2 bullets after creation, got %d", bm.GetBulletCount())
	}

	// Scenario 1: Bullet out of bounds
	bullet1.SetOutOfBounds(true)

	// Scenario 2: Bullet in collision
	enemy := &mocks.MockEnemy{}
	bullet2.OnCollision(enemy)

	// Process destruction - bullets should destroy themselves and publish events
	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	// Check events
	events := eventManager.GetPublishedEvents()
	destroyedEvents := 0
	for _, e := range events {
		if e.Type == interfaces.BulletDestroyed {
			destroyedEvents++
		}
	}
	if destroyedEvents < 2 {
		t.Errorf("Expected at least 2 BulletDestroyed events, got %d", destroyedEvents)
	}

	// Process the destruction events
	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("Second Update returned an error: %v", err)
	}

	// Check that bullets were destroyed
	finalCount := bm.GetBulletCount()
	if finalCount != 0 {
		t.Errorf("Expected 0 bullets after destruction, got %d", finalCount)
	}
}

func TestBulletManagerHandleBulletCreated(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	if bm.GetBulletCount() != 0 {
		t.Errorf("Expected 0 bullets initially, got %d", bm.GetBulletCount())
	}

	bullet := mocks.NewMockBullet(100, 100, false, eventManager)
	eventManager.Publish(interfaces.BulletCreated, bullet)

	err = bm.Update(0.16)
	if err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if bm.GetBulletCount() != 1 {
		t.Errorf("Expected 1 bullet after BulletCreated event, got %d", bm.GetBulletCount())
	}

	// Check that the added bullet is the one we created
	bullets := bm.GetBullets()
	if len(bullets) > 0 && bullets[0] != bullet {
		t.Error("The added bullet is not the one we created")
	}
}

func TestBulletManagerGetRenderableEntities(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx := context.Background()

	bm.Initialize(ctx)

	bullet1 := mocks.NewMockBullet(100, 100, false, eventManager)
	bullet2 := mocks.NewMockBullet(200, 200, true, eventManager)
	bullet2.Health = 0 // Make it dead

	eventManager.Publish(interfaces.BulletCreated, bullet1)
	eventManager.Publish(interfaces.BulletCreated, bullet2)
	bm.Update(0.16) // Process events

	renderables := bm.GetRenderableEntities()

	// Only alive bullets should be rendered
	if len(renderables) != 1 {
		t.Errorf("Expected 1 renderable bullet, got %d", len(renderables))
	}
}

func TestBulletManagerShutdown(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	bm := NewBulletManager(eventManager)
	ctx, cancel := context.WithCancel(context.Background())

	err := bm.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize BulletManager: %v", err)
	}

	bullet := mocks.NewMockBullet(100, 100, false, eventManager)
	eventManager.Publish(interfaces.BulletCreated, bullet)
	bm.Update(0.16) // Process event

	if bm.GetBulletCount() != 1 {
		t.Fatalf("Expected 1 bullet before shutdown, got %d", bm.GetBulletCount())
	}

	cancel()
	bm.Shutdown()

	if bm.GetBulletCount() != 0 {
		t.Error("Shutdown did not clear bullets")
	}

	err = bm.Update(0.16)
	if err == nil {
		t.Error("Update after shutdown should return an error")
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
