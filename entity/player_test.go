package entity

import (
	"context"
	"testing"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
	"github.com/ajkula/shmup/types"
)

func TestNewPlayer(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	player := NewPlayer(types.Vector2D{X: 100, Y: 200}, eventManager)

	if player.GetPosition().X != 100 || player.GetPosition().Y != 200 {
		t.Errorf("NewPlayer position: got (%v,%v), want (100,200)", player.GetPosition().X, player.GetPosition().Y)
	}
	width, height := player.GetSize()
	if width != 32 || height != 32 {
		t.Errorf("NewPlayer size: got (%v,%v), want (32,32)", width, height)
	}
	if player.Speed != 5 {
		t.Errorf("NewPlayer speed: got %v, want 5", player.Speed)
	}
	if player.GetHealth() != 100 {
		t.Errorf("NewPlayer health: got %v, want 100", player.GetHealth())
	}
	if player.ShootCooldown != 0 {
		t.Errorf("NewPlayer ShootCooldown: got %v, want 0", player.ShootCooldown)
	}
}

func TestPlayerUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)

	err := player.Update(0.1)
	if err != nil {
		t.Errorf("Player.Update() returned an error: %v", err)
	}

	if player.ShootCooldown != -0.1 {
		t.Errorf("Player.ShootCooldown after update: got %v, want -0.1", player.ShootCooldown)
	}
}

func TestPlayerCanCollideWith(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)

	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	if !player.CanCollideWith(enemy) {
		t.Error("Player should be able to collide with Enemy")
	}

	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)
	if !player.CanCollideWith(boss) {
		t.Error("Player should be able to collide with Boss")
	}

	friendlyBullet := NewBullet(100, 100, false, eventManager)
	if player.CanCollideWith(friendlyBullet) {
		t.Error("Player should not be able to collide with friendly Bullet")
	}

	enemyBullet := NewBullet(100, 100, true, eventManager)
	if !player.CanCollideWith(enemyBullet) {
		t.Error("Player should be able to collide with enemy Bullet")
	}
}

func TestPlayerOnCollision(t *testing.T) {
	ctx := context.Background()
	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)
	damagedEvents, err := eventManager.Subscribe(interfaces.PlayerDamaged)
	if err != nil {
		t.Errorf("Error happened during Subscription to PlayerDamaged")
	}

	player.OnCollision(nil)
	eventManager.Update(0)

	if player.GetHealth() != 90 {
		t.Errorf("Player health after collision: got %v, want 90", player.GetHealth())
	}

	select {
	case e := <-damagedEvents:
		if e.Type != interfaces.PlayerDamaged {
			t.Errorf("Expected PlayerDamaged event, got %v", e.Type)
		}
	default:
		t.Error("No PlayerDamaged event received")
	}
}

func TestPlayerShoot(t *testing.T) {
	ctx := context.Background()
	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)
	shotEvents, err := eventManager.Subscribe(interfaces.PlayerShot)
	if err != nil {
		t.Errorf("Error happened during Subscription to PlayerShot")
	}

	player.Shoot()
	eventManager.Update(0)

	if player.ShootCooldown != 0.2 {
		t.Errorf("Player ShootCooldown after shooting: got %v, want 0.2", player.ShootCooldown)
	}

	select {
	case e := <-shotEvents:
		if e.Type != interfaces.PlayerShot {
			t.Errorf("Expected PlayerShot event, got %v", e.Type)
		}
	default:
		t.Error("No PlayerShot event received")
	}

	player.Shoot()
	eventManager.Update(0) // process any events
	select {
	case <-shotEvents:
		t.Error("Player should not be able to shoot during cooldown")
	default:
		// good one
	}
}

func TestPlayerCanShoot(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)

	if !player.CanShoot() {
		t.Error("Player should be able to shoot initially")
	}

	player.Shoot()

	if player.CanShoot() {
		t.Error("Player should not be able to shoot immediately after shooting")
	}

	player.Update(0.3)

	if !player.CanShoot() {
		t.Error("Player should be able to shoot after cooldown")
	}
}
