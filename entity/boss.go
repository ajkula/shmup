package entity

import (
	"time"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Boss struct {
	types.BaseEntity
	phase         int
	eventManager  interfaces.EventManagerInterface
	shootCooldown *ThreadSafeCooldown
}

func NewBoss(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Boss {
	return &Boss{
		BaseEntity: types.BaseEntity{
			Position: position,
			Width:    64, Height: 64,
			Speed:  1,
			Health: 1000,
		},
		phase:         1,
		eventManager:  eventManager,
		shootCooldown: NewThreadSafeCooldown(0.2), // 0.2 seconds like original
	}
}

func (b *Boss) Update(deltaTime float64) error {
	b.shootCooldown.Update(deltaTime) // Update cooldown with deltaTime
	if b.CanShoot() {
		b.Shoot()
	}
	if b.Health <= 500 && b.phase == 1 {
		b.ChangePhase(2)
	}
	return nil
}

func (b *Boss) Draw(screen *ebiten.Image) {
	// todo
}

func (b *Boss) CanCollideWith(other types.Entity) bool {
	switch o := other.(type) {
	case *Player:
		return true
	case *Bullet:
		return !o.IsEnemyBullet()
	default:
		return false
	}
}

func (b *Boss) OnCollision(other types.Entity) {
	b.TakeDamage(10)
	b.eventManager.Publish(interfaces.BossDamaged, b)
	if b.Health <= 0 {
		b.eventManager.Publish(interfaces.BossDefeated, b)
	}
}

func (b *Boss) CanShoot() bool {
	return b.shootCooldown.CanAct()
}

func (b *Boss) Shoot() bool {
	if b.shootCooldown.TryAct() {
		b.eventManager.Publish(interfaces.BossShot, b)
		return true
	}
	return false
}

func (b *Boss) ChangePhase(newPhase int) {
	b.phase = newPhase
	b.eventManager.Publish(interfaces.BossPhaseChanged, b)
}

func (b *Boss) GetShootCooldownRemaining() time.Duration {
	return b.shootCooldown.GetRemainingCooldown()
}

func (b *Boss) ResetShootCooldown() {
	b.shootCooldown.Reset()
}

func (b *Boss) SetShootRate(ratePerSecond float64) {
	cooldownTime := 1.0 / ratePerSecond // Convert to seconds
	b.shootCooldown.SetCooldownTime(cooldownTime)
}

var _ types.GameEntity = (*Boss)(nil)
