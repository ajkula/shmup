package entity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ajkula/shmup/graphics"
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

	// Sprite system: 7 pixels * 4px = 28, 7 pixels * 4px = 28
	if enemy.Width != 28 || enemy.Height != 28 {
		t.Errorf("NewEnemy size: got (%v,%v), want (28,28)", enemy.Width, enemy.Height)
	}

	// Scout Level1 stats: Speed = 3.0
	if enemy.Speed != 3.0 {
		t.Errorf("NewEnemy speed: got %v, want 3.0", enemy.Speed)
	}

	// Scout Level1 stats: HP = 1
	if enemy.Health != 1 {
		t.Errorf("NewEnemy health: got %v, want 1", enemy.Health)
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

	// Updated: All bullets can collide with enemies now
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

	// Test collision with bullet (should damage but not destroy with 1 HP enemy)
	bullet := NewBullet(100, 100, false, eventManager)
	enemy.OnCollision(bullet)

	select {
	case e := <-destroyedEvents:
		if e.Type != interfaces.EnemyDestroyed {
			t.Errorf("Expected EnemyDestroyed event, got %v", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
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
	select {
	case <-shotEvents:
		t.Error("Enemy should not be able to shoot during cooldown")
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	// Scout Level1 FireRate = 2.0, so need 2.0/0.1 = 20 updates to clear cooldown
	for i := 0; i < 20; i++ {
		enemy.Update(testDeltaTime)
	}

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
	ctx := context.Background()
	eventManager := mocks.NewMockEventManager()
	eventManager.Initialize(ctx)

	t.Run("Initial and cooldown", func(t *testing.T) {
		enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
		shotEvents, _ := eventManager.Subscribe(interfaces.EnemyShot)

		enemy.Update(0.01)
		select {
		case <-shotEvents:
		case <-time.After(10 * time.Millisecond):
			t.Error("Initial shot expected")
		}

		enemy.Update(0.01)
		select {
		case <-shotEvents:
			t.Error("Should not shoot during cooldown")
		case <-time.After(10 * time.Millisecond):
		}

		stats := graphics.GetEnemyStats(graphics.Scout, graphics.Level1)
		for i := 0.0; i < stats.FireRate; i += 0.1 {
			enemy.Update(0.1)
		}

		select {
		case <-shotEvents:
		case <-time.After(10 * time.Millisecond):
			t.Error("Shot expected after cooldown")
		}
	})

	for i := 1; i <= 3; i++ {
		t.Run(fmt.Sprintf("Regular shot %d", i), func(t *testing.T) {
			enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
			shotEvents, _ := eventManager.Subscribe(interfaces.EnemyShot)

			enemy.Update(0.01)

			select {
			case <-shotEvents:
			case <-time.After(10 * time.Millisecond):
				t.Error("Expected shot, but none fired")
			}
		})
	}
}

func TestEnemyMovement(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	startPosition := types.Vector2D{X: 100, Y: 100}
	enemy := NewEnemy(startPosition, eventManager)

	enemy.Update(testDeltaTime)

	// Enemy now moves down automatically: Y += Speed * deltaTime * 60
	// Scout Level1 Speed = 3.0, so Y += 3.0 * 0.1 * 60 = 18
	expectedPosition := types.Vector2D{X: 100, Y: 118}
	if enemy.Position != expectedPosition {
		t.Errorf("Enemy position after update: got %v, want %v", enemy.Position, expectedPosition)
	}
}

func TestEnemyTypes(t *testing.T) {
	eventManager := mocks.NewMockEventManager()

	// Test different enemy types
	scout := NewEnemyWithType(types.Vector2D{X: 0, Y: 0}, eventManager, graphics.Scout, graphics.Level1)
	if scout.GetEnemyType() != graphics.Scout {
		t.Error("Scout enemy type not set correctly")
	}

	fighter := NewEnemyWithType(types.Vector2D{X: 0, Y: 0}, eventManager, graphics.Fighter, graphics.Level1)
	if fighter.GetEnemyType() != graphics.Fighter {
		t.Error("Fighter enemy type not set correctly")
	}

	heavy := NewEnemyWithType(types.Vector2D{X: 0, Y: 0}, eventManager, graphics.Heavy, graphics.Level1)
	if heavy.GetEnemyType() != graphics.Heavy {
		t.Error("Heavy enemy type not set correctly")
	}
}

func TestEnemyUpgrade(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	enemy := NewEnemyWithType(types.Vector2D{X: 0, Y: 0}, eventManager, graphics.Scout, graphics.Level1)

	originalSpeed := enemy.Speed
	originalHealth := enemy.Health

	enemy.UpgradeLevel()

	if enemy.GetEnemyLevel() != graphics.Level2 {
		t.Error("Enemy level not upgraded correctly")
	}

	// Scout Level2 should have higher speed and HP
	if enemy.Speed <= originalSpeed {
		t.Error("Enemy speed should increase after upgrade")
	}

	if enemy.Health <= originalHealth {
		t.Error("Enemy health should increase after upgrade")
	}
}
