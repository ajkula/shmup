package entity

import (
	"context"
	"testing"
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/mocks"
	"github.com/ajkula/shmup/types"
)

func TestNewBoss(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	boss := NewBoss(types.Vector2D{X: 100, Y: 200}, eventManager)

	if boss.Position.X != 100 || boss.Position.Y != 200 {
		t.Errorf("NewBoss position: got (%v,%v), want (100,200)", boss.Position.X, boss.Position.Y)
	}
	if boss.Width != 64 || boss.Height != 64 {
		t.Errorf("NewBoss size: got (%v,%v), want (64,64)", boss.Width, boss.Height)
	}
	if boss.Speed != 1 {
		t.Errorf("NewBoss speed: got %v, want 1", boss.Speed)
	}
	if boss.Health != 1000 {
		t.Errorf("NewBoss health: got %v, want 1000", boss.Health)
	}
	if boss.ShootCooldown != 0 {
		t.Errorf("NewBoss shootCooldown: got %v, want 0", boss.ShootCooldown)
	}
}

func TestBossUpdate(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)

	boss.ShootCooldown = 0.2
	err := boss.Update(0.1)
	if err != nil {
		t.Errorf("Boss.Update() returned an error: %v", err)
	}

	if boss.ShootCooldown != 0.1 {
		t.Errorf("Boss.ShootCooldown after update: got %v, want 0.1", boss.ShootCooldown)
	}
}

func TestBossCanCollideWith(t *testing.T) {
	eventManager := mocks.NewMockEventManager()
	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)

	player := NewPlayer(types.Vector2D{X: 100, Y: 100}, eventManager)
	if !boss.CanCollideWith(player) {
		t.Error("Boss should be able to collide with Player")
	}

	enemy := NewEnemy(types.Vector2D{X: 100, Y: 100}, eventManager)
	if boss.CanCollideWith(enemy) {
		t.Error("Boss should not be able to collide with Enemy")
	}

	friendlyBullet := NewBullet(0, 0, false, eventManager)
	if !boss.CanCollideWith(friendlyBullet) {
		t.Error("Boss should be able to collide with player Bullet")
	}

	enemyBullet := NewBullet(0, 0, true, eventManager)
	if boss.CanCollideWith(enemyBullet) {
		t.Error("Boss should not be able to collide with enemy Bullet")
	}
}

func TestBossOnCollision(t *testing.T) {
	ctx := context.Background()
	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)
	boss.Health = 1000

	damagedEvents, err := eventManager.Subscribe(interfaces.BossDamaged)
	if err != nil {
		t.Errorf("Error happened during Subscription to BossDamager")
	}

	go eventManager.Run(ctx)

	boss.OnCollision(nil)

	if boss.Health != 990 {
		t.Errorf("Boss health after collision: got %v, want 990", boss.Health)
	}

	select {
	case e := <-damagedEvents:
		if e.Type != interfaces.BossDamaged {
			t.Errorf("Expected BossDamaged event, got %v", e.Type)
		}
	case <-time.After(time.Millisecond * 100):
		t.Error("No BossDamaged event received")
	}
}

func TestBossShoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)
	shotEvents, err := eventManager.Subscribe(interfaces.BossShot)
	if err != nil {
		t.Errorf("Error happened during Subscription to BossShot")
	}

	go eventManager.Run(ctx)

	// shoot initial
	boss.Update(0.1)

	select {
	case e := <-shotEvents:
		if e.Type != interfaces.BossShot {
			t.Errorf("Expected BossShot event, got %v", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("No BossShot event received")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	// ne peut pas tirer
	boss.Update(0.1)
	select {
	case <-shotEvents:
		t.Error("Boss should not be able to shoot during cooldown")
	case <-time.After(100 * time.Millisecond):
		// comportement attendu
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	// attend que le cooldown soit terminé et vérifie que le boss peut tirer à nouveau
	time.Sleep(200 * time.Millisecond)
	boss.Update(0.1)

	select {
	case e := <-shotEvents:
		if e.Type != interfaces.BossShot {
			t.Errorf("Expected BossShot event, got %v", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("No BossShot event received after cooldown")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}
}

func TestBossAutoShoot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)
	shotEvents, err := eventManager.Subscribe(interfaces.BossShot)
	if err != nil {
		t.Errorf("Error happened during Subscription to BossShot")
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
		{"After cooldown", 0.19, true, 0},
		{"Regular shot 1", boss.maxCooldown + 0.01, true, 10 * time.Millisecond},
		{"Regular shot 2", boss.maxCooldown + 0.01, true, 10 * time.Millisecond},
		{"Regular shot 3", boss.maxCooldown + 0.01, true, 10 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			boss.Update(tc.updateTime)
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

func TestBossPhaseChange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	eventManager := mocks.NewMockEventManager()
	err := eventManager.Initialize(ctx)
	if err != nil {
		t.Fatalf("Failed to initialize EventManager: %v", err)
	}

	boss := NewBoss(types.Vector2D{X: 100, Y: 100}, eventManager)
	phaseChangeEvents, err := eventManager.Subscribe(interfaces.BossPhaseChanged)
	if err != nil {
		t.Errorf("Error happened during Subscription to BossPhaseChanged")
	}

	go eventManager.Run(ctx)

	initialHealth := boss.Health
	damagePerShot := 10
	shotsNeeded := (initialHealth-500)/damagePerShot + 1

	for i := 1; i < shotsNeeded; i++ {
		boss.OnCollision(nil) // hit
		boss.Update(0.1)      // shoot > boss update

		if i < shotsNeeded-1 {
			select {
			case <-phaseChangeEvents:
				t.Errorf("Boss changed phase too early at shot %d", i+1)
			case <-time.After(10 * time.Millisecond):
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		}
	}

	select {
	case e := <-phaseChangeEvents:
		if e.Type != interfaces.BossPhaseChanged {
			t.Errorf("Expected BossPhaseChanged event, got %v", e.Type)
		}
	case <-time.After(time.Second):
		t.Error("No BossPhaseChanged event received")
	case <-ctx.Done():
		t.Fatal("Test timed out")
	}

	if boss.phase != 2 {
		t.Errorf("Expected boss to be in phase 2, but got phase %d", boss.phase)
	}

	if boss.Health > 500 {
		t.Errorf("Expected boss health to be <= 500, but got %d", boss.Health)
	}
}
