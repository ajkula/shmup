package entity

import (
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	types.BaseEntity
	ShootCooldown float64
	eventManager  interfaces.EventManagerInterface
}

func NewPlayer(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Player {
	return &Player{
		BaseEntity: types.BaseEntity{
			Position: position,
			Width:    32, Height: 32,
			Speed:  5,
			Health: 100,
		},
		ShootCooldown: 0,
		eventManager:  eventManager,
	}
}

func (p *Player) Update(deltaTime float64) error {
	p.ShootCooldown -= deltaTime
	return nil
}

func (p *Player) Draw(screen *ebiten.Image) {
	// TODO
}

func (p *Player) CanCollideWith(other types.Entity) bool {
	switch o := other.(type) {
	case *Enemy:
		return true
	case *Boss:
		return true
	case *Bullet:
		return o.IsEnemyBullet()
	default:
		return false
	}
}

func (p *Player) OnCollision(other types.Entity) {
	p.TakeDamage(10)
	p.eventManager.Publish(interfaces.PlayerDamaged, p)
	if p.Health <= 0 {
		p.eventManager.Publish(interfaces.PlayerDestroyed, p)
	}
}

func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0
}

func (p *Player) Shoot() {
	if p.CanShoot() {
		p.eventManager.Publish(interfaces.PlayerShot, p)
		p.ShootCooldown = 0.2
	}
}

var _ types.GameEntity = (*Player)(nil)
