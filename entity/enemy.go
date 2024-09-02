package entity

import (
	"math"

	"github.com/ajkula/shmup/common"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	types.BaseEntity
	shootCooldown float64
	maxCooldown   float64
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
		shootCooldown: 0,
		maxCooldown:   1.0,
		eventManager:  eventManager,
	}
}

func (e *Enemy) Update(deltaTime float64) error {
	e.shootCooldown = math.Max(0, e.shootCooldown-deltaTime)
	if e.CanShoot() {
		e.Shoot()
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
	if e.Health <= 0 {
		e.eventManager.Publish(interfaces.EnemyDestroyed, e)
	}
}

func (e *Enemy) CanShoot() bool {
	return e.shootCooldown <= common.Epsilon
}

func (e *Enemy) Shoot() {
	e.eventManager.Publish(interfaces.EnemyShot, e)
	e.shootCooldown = e.maxCooldown
}

var _ types.GameEntity = (*Enemy)(nil)
