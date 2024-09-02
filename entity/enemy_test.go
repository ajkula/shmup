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
	if enemy.shootCooldown != 0 {
		t.Errorf("NewEnemy shootCooldown: got %v, want 0", enemy.shootCooldown)
	}
	if enemy.maxCooldown != 1.0 {
		t.Errorf("NewEnemy maxCooldown: got %v, want 1.0", enemy.maxCooldown)
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

	if enemy.shootCooldown != enemy.maxCooldown {
		t.Errorf("Enemy shootCooldown after update: got %v, want %v", enemy.shootCooldown, enemy.maxCooldown)
	}

	enemy.Update(testDeltaTime)
	expectedCooldown := enemy.maxCooldown - testDeltaTime
	if enemy.shootCooldown != expectedCooldown {
		t.Errorf("Enemy shootCooldown after second update: got %v, want %v", enemy.shootCooldown, expectedCooldown)
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
		t.Errorf("Error happened during Subscription to EnemyDestroyed")
	}

	go eventManager.Run(ctx)

	enemy.OnCollision(nil)

	if enemy.Health != 10 {
		t.Errorf("Enemy health after collision: got %v, want 10", enemy.Health)
	}

	select {
	case <-destroyedEvents:
		t.Error("EnemyDestroyed event received too early")
	case <-time.After(100 * time.Millisecond):
		// good one
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	enemy.OnCollision(nil)

	select {
	case e := <-destroyedEvents:
		if e.Type != interfaces.EnemyDestroyed {
			t.Errorf("Expected EnemyDestroyed event, got %v", e.Type)
		}
	case <-time.After(time.Second):
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
	select {
	case <-shotEvents:
		t.Error("Enemy should not be able to shoot during cooldown")
	case <-time.After(100 * time.Millisecond):
		// good one
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	for i := 0; i < int(enemy.maxCooldown/testDeltaTime); i++ {
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
		name           string
		updateTime     float64
		expectShot     bool
		additionalWait time.Duration
	}{
		{"Initial shot", 0.01, true, 0},
		{"During cooldown", 0.01, false, 0},
		{"After cooldown", 0.99, true, 0},
		{"Regular shot 1", enemy.maxCooldown + 0.01, true, 10 * time.Millisecond},
		{"Regular shot 2", enemy.maxCooldown + 0.01, true, 10 * time.Millisecond},
		{"Regular shot 3", enemy.maxCooldown + 0.01, true, 10 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enemy.Update(tc.updateTime)
			time.Sleep(tc.additionalWait)

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

	// expectedPosition := startPosition.Add(types.Vector2D{X: 0, Y: 1}.Multiply(enemy.Speed * testDeltaTime))
	expectedPosition := types.Vector2D{X: 100, Y: 100}
	if enemy.Position != expectedPosition {
		t.Errorf("Enemy position after update: got %v, want %v", enemy.Position, expectedPosition)
	}
}
