package entity

import (
	"image/color"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	types.BaseEntity
	eventManager  interfaces.EventManagerInterface
	ShootCooldown float64

	EnemyType     graphics.EnemyType
	EnemyLevel    graphics.EnemyLevel
	sprite        *graphics.SpriteData
	bulletPattern interfaces.BulletPattern
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
			Color:    color.RGBA{255, 0, 0, 255},
		},
		eventManager:  eventManager,
		ShootCooldown: 0,
		EnemyType:     enemyType,
		EnemyLevel:    level,
		sprite:        sprite,
		bulletPattern: selectEnemyPattern(enemyType, level),
	}

	return enemy
}

func selectEnemyPattern(enemyType graphics.EnemyType, level graphics.EnemyLevel) interfaces.BulletPattern {
	switch enemyType {
	case graphics.Scout:
		switch level {
		case graphics.Level1:
			return NewStreamPattern(200)
		case graphics.Level2:
			return NewSprayPattern(3, 30, 180)
		default:
			return NewBurstPattern(5, 150, 30)
		}
	case graphics.Fighter:
		switch level {
		case graphics.Level1:
			return NewSprayPattern(3, 45, 200)
		case graphics.Level2:
			return NewCirclePattern(8, 150)
		default:
			return NewSpiralPattern(6, 180, 2.0)
		}
	case graphics.Heavy:
		switch level {
		case graphics.Level1:
			return NewBurstPattern(5, 160, 20)
		case graphics.Level2:
			return NewWavePattern(7, 140, 200)
		default:
			return NewCirclePattern(16, 120)
		}
	default:
		return NewStreamPattern(200)
	}
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
	switch other := other.(type) {
	case *Player:
		return true
	case *Bullet:
		return !other.IsEnemyBullet()
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
	return e.ShootCooldown <= 0 &&
		e.Position.Y > 0 &&
		e.Position.Y < float64(config.Config.ScreenHeight*3/4)
}

func (e *Enemy) Shoot() {
	if e.CanShoot() {
		stats := graphics.GetEnemyStats(e.EnemyType, e.EnemyLevel)

		patternEvent := interfaces.PatternShootEvent{
			Shooter: e,
			Pattern: e.bulletPattern,
		}

		e.eventManager.Publish(interfaces.EnemyShot, patternEvent)
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
		e.bulletPattern = selectEnemyPattern(e.EnemyType, e.EnemyLevel)
	}
}

var _ types.GameEntity = (*Enemy)(nil)
