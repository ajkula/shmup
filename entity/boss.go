package entity

import (
	"math"

	"github.com/ajkula/shmup/common"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Boss struct {
	types.BaseEntity
	phase         int
	eventManager  interfaces.EventManagerInterface
	ShootCooldown float64
	maxCooldown   float64
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
		ShootCooldown: 0,
		maxCooldown:   0.2,
	}
}

func (b *Boss) Update(deltaTime float64) error {
	b.ShootCooldown = math.Max(0, b.ShootCooldown-deltaTime)
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
	return b.ShootCooldown < common.Epsilon
}

func (b *Boss) Shoot() {
	b.eventManager.Publish(interfaces.BossShot, b)
	b.ShootCooldown = b.maxCooldown
}

func (b *Boss) ChangePhase(newPhase int) {
	b.phase = newPhase
	b.eventManager.Publish(interfaces.BossPhaseChanged, b)
}

var _ types.GameEntity = (*Boss)(nil)
