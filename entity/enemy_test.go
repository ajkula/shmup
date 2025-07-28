package entity

import (
	"context"
	"testing"
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
	"github.com/ajkula/shmup/types"
)

const testDeltaTime = 0.1

func TestNewEnemy(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	enemy := NewEnemy(types.Vector2D{X: 100, Y: 200}, eventManager)

	if enemy.Position.X != 100 || enemy.Position.Y != 200 {
		t.Errorf("NewEnemy position: got (%v,%v), want (100,200)", enemy.Position.X, enemy.Position.Y)
	}
	if enemy.Width != 32 || enemy.Height != 32 {
		t.Errorf("NewEnemy size: got (%v,%v), want (32,32)", enemy.Width, enemy.Height)
	}
	if enemy.Speed != 2 {
		t.Errorf("NewEnemy speed: got %v, want 2", enemy.Speed)
	}
	if enemy.Health != 20 {
		t.Errorf("NewEnemy health: got %v, want 20", enemy.Health)
	}
	if !enemy.CanShoot() {
		t.Error("NewEnemy should be able to shoot initially")
	}
}

func TestEnemyUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)

	if !enemy.CanShoot() {
		t.Error("Enemy should be able to shoot initially")
	}

	err := enemy.Update(testDeltaTime)
	if err != nil {
		t.Errorf("Enemy.Update() returned an error: %v", err)
	}

	if enemy.CanShoot() {
		t.Error("Enemy should not be able to shoot immediately after shooting")
	}

	remaining := enemy.GetShootCooldownRemaining()
	if remaining <= 0 {
		t.Error("Enemy should have cooldown remaining after shooting")
	}
}

func TestEnemyCanCollideWith(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)

	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)
	if !enemy.CanCollideWith(player) {
		t.Error("Enemy should be able to collide with Player")
	}

	otherEnemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	if enemy.CanCollideWith(otherEnemy) {
		t.Error("Enemy should not be able to collide with another Enemy")
	}

	friendlyBullet := NewBullet(0, 0, false, eventManager)
	if !enemy.CanCollideWith(friendlyBullet) {
		t.Error("Enemy should be able to collide with player Bullet")
	}

	enemyBullet := NewBullet(0, 0, true, eventManager)
	if enemy.CanCollideWith(enemyBullet) {
		t.Error("Enemy should not be able to collide with enemy Bullet")
	}
}

func TestEnemyOnCollision(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	destroyedEvents, err := eventManager.Subscribe(interfaces.EnemyDestroyed)
	if err != nil {
		t.Errorf("Error during Subscription to EnemyDestroyed")
	}

	go eventManager.Run(ctx)

	time.Sleep(20 * time.Millisecond)

	enemy.OnCollision(nil)

	select {
	case <-destroyedEvents:
		t.Error("EnemyDestroyed event received too early")
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	enemy.OnCollision(nil)
	enemy.Update(0.1) // ← FIX: Update needed to publish EnemyDestroyed

	select {
	case e := <-destroyedEvents:
		if e.Type != interfaces.EnemyDestroyed {
			t.Errorf("Expected EnemyDestroyed event, got %v", e.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("No EnemyDestroyed event received")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}
}

func TestEnemyShoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	shotEvents, err := eventManager.Subscribe(interfaces.EnemyShot)
	if err != nil {
		t.Errorf("Error happened during Subscription to EnemyShot")
	}

	go eventManager.Run(ctx)

	enemy.Update(testDeltaTime)

	select {
	case e := <-shotEvents:
		if e.Type != interfaces.EnemyShot {
			t.Errorf("Expected EnemyShot event, got %v", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("No EnemyShot event received")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	enemy.Update(testDeltaTime)
	time.Sleep(10 * time.Millisecond)
	select {
	case <-shotEvents:
		t.Error("Enemy should not be able to shoot during cooldown")
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	for i := 0; i < 10; i++ {
		enemy.Update(testDeltaTime)
	}
	time.Sleep(10 * time.Millisecond)

	select {
	case e := <-shotEvents:
		if e.Type != interfaces.EnemyShot {
			t.Errorf("Expected EnemyShot event, got %v", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("No EnemyShot event received after cooldown")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}
}

func TestEnemyAutoShoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	shotEvents, err := eventManager.Subscribe(interfaces.EnemyShot)
	if err != nil {
		t.Errorf("Error happened during Subscription to EnemyShot")
	}

	go eventManager.Run(ctx)

	testCases := []struct {
		name       string
		updateTime float64
		expectShot bool
		updates    int
	}{
		{"Initial shot", 0.01, true, 1},
		{"During cooldown", 0.01, false, 1},
		{"After cooldown", 0.1, true, 10}, // 10 * 0.1 = 1.0 second
		{"Regular shot 1", 0.1, true, 11}, // +1.1 seconds
		{"Regular shot 2", 0.1, true, 11},
		{"Regular shot 3", 0.1, true, 11},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < tc.updates; i++ {
				enemy.Update(tc.updateTime)
			}
			time.Sleep(10 * time.Millisecond)

			select {
			case <-shotEvents:
				if !tc.expectShot {
					t.Error("Unexpected shot fired")
				}
			case <-time.After(100 * time.Millisecond):
				if tc.expectShot {
					t.Error("Expected shot, but none fired")
				}
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		})
	}
}

func TestEnemyMovement(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	startPosition := types.Vector2D{X: 100, Y: 100}
	enemy := NewEnemy(startPosition, eventManager)

	enemy.Update(testDeltaTime)

	expectedPosition := types.Vector2D{X: 100, Y: 100}
	if enemy.Position != expectedPosition {
		t.Errorf("Enemy position after update: got %v, want %v", enemy.Position, expectedPosition)
	}
}
