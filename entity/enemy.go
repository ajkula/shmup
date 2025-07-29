package entity

import (
	"image/color"

	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	types.BaseEntity
	eventManager  interfaces.EventManagerInterface
	ShootCooldown float64

	EnemyType  graphics.EnemyType
	EnemyLevel graphics.EnemyLevel
	sprite     *graphics.SpriteData
}

func NewEnemy(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Enemy {
	return NewEnemyWithType(position, eventManager, graphics.Scout, graphics.Level1)
}

func NewEnemyWithType(position types.Vector2D, eventManager interfaces.EventManagerInterface,
	enemyType graphics.EnemyType, level graphics.EnemyLevel) *Enemy {

	stats := graphics.GetEnemyStats(enemyType, level)
	sprite := graphics.GetEnemySprite(enemyType, level)

	enemy := &Enemy{
		BaseEntity: types.BaseEntity{
			Position: position,
			Width:    float64(sprite.Width * 4),
			Height:   float64(sprite.Height * 4),
			Speed:    stats.Speed,
			Health:   stats.HP,
			Color:    color.RGBA{255, 0, 0, 255}, // Not used with sprite system
		},
		eventManager:  eventManager,
		ShootCooldown: 0,
		EnemyType:     enemyType,
		EnemyLevel:    level,
		sprite:        sprite,
	}

	return enemy
}

func (e *Enemy) Update(deltaTime float64) error {
	e.ShootCooldown -= deltaTime
	if e.ShootCooldown < 0 {
		e.ShootCooldown = 0
	}

	e.Position.Y += e.Speed * deltaTime * 60

	if e.CanShoot() {
		e.Shoot()
	}

	return nil
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	graphics.DrawSprite(screen, e.sprite, e.Position.X, e.Position.Y)
}

func (e *Enemy) CanCollideWith(other types.Entity) bool {
	switch other.(type) {
	case *Player:
		return true
	case *Bullet:
		return true
	}
	return false
}

func (e *Enemy) OnCollision(other types.Entity) {
	switch other.(type) {
	case *Player:
		e.TakeDamage(e.Health)
		e.eventManager.Publish(interfaces.EnemyDestroyed, e)
	case *Bullet:
		e.TakeDamage(10)
		if e.Health <= 0 {
			e.eventManager.Publish(interfaces.EnemyDestroyed, e)
		} else {
			e.eventManager.Publish(interfaces.EnemyDamaged, e)
		}
	}
}

func (e *Enemy) CanShoot() bool {
	return e.ShootCooldown <= 0
}

func (e *Enemy) Shoot() {
	if e.CanShoot() {
		stats := graphics.GetEnemyStats(e.EnemyType, e.EnemyLevel)
		e.eventManager.Publish(interfaces.EnemyShot, e)
		e.ShootCooldown = stats.FireRate
	}
}

func (e *Enemy) GetShootCooldownRemaining() float64 {
	return e.ShootCooldown
}

func (e *Enemy) GetEnemyType() graphics.EnemyType {
	return e.EnemyType
}

func (e *Enemy) GetEnemyLevel() graphics.EnemyLevel {
	return e.EnemyLevel
}

func (e *Enemy) UpgradeLevel() {
	if e.EnemyLevel < graphics.Level5 {
		e.EnemyLevel++
		e.sprite = graphics.GetEnemySprite(e.EnemyType, e.EnemyLevel)

		stats := graphics.GetEnemyStats(e.EnemyType, e.EnemyLevel)
		e.Speed = stats.Speed
		e.Health = stats.HP
	}
}

var _ types.GameEntity = (*Enemy)(nil)
