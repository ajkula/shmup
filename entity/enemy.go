package entity

import (
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	types.BaseEntity
	shootCooldown *ThreadSafeCooldown
	eventManager  interfaces.EventManagerInterface
}

func NewEnemy(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Enemy {
	return &Enemy{
		BaseEntity: types.BaseEntity{
			Position: position,
			Width:    32, Height: 32,
			Speed:  2,
			Health: 20,
		},
		shootCooldown: NewThreadSafeCooldown(1.0), // 1.0 second like original
		eventManager:  eventManager,
	}
}

func (e *Enemy) Update(deltaTime float64) error {
	e.shootCooldown.Update(deltaTime) // Update cooldown with deltaTime
	if e.CanShoot() {
		e.Shoot()
	}
	if e.Health <= 0 {
		e.eventManager.Publish(interfaces.EnemyDestroyed, e)
	}
	return nil
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	// todo
}

func (e *Enemy) CanCollideWith(other types.Entity) bool {
	switch o := other.(type) {
	case *Player:
		return true
	case *Bullet:
		return !o.IsEnemyBullet()
	default:
		return false
	}
}

func (e *Enemy) OnCollision(other types.Entity) {
	e.TakeDamage(10)
	e.eventManager.Publish(interfaces.EnemyDamaged, e)
}

func (e *Enemy) CanShoot() bool {
	return e.shootCooldown.CanAct()
}

func (e *Enemy) Shoot() bool {
	if e.shootCooldown.TryAct() {
		e.eventManager.Publish(interfaces.EnemyShot, e)
		return true
	}
	return false
}

func (e *Enemy) GetShootCooldownRemaining() time.Duration {
	return e.shootCooldown.GetRemainingCooldown()
}

func (e *Enemy) ResetShootCooldown() {
	e.shootCooldown.Reset()
}

func (e *Enemy) SetShootRate(ratePerSecond float64) {
	cooldownTime := 1.0 / ratePerSecond // Convert to seconds
	e.shootCooldown.SetCooldownTime(cooldownTime)
}

var _ types.GameEntity = (*Enemy)(nil)
