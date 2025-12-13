package entity

import (
	"time"

	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Boss struct {
	types.BaseEntity
	phase           int
	eventManager    interfaces.EventManagerInterface
	shootCooldown   *ThreadSafeCooldown
	patternSequence []interfaces.BulletPattern
	currentPattern  int
	patternTimer    float64
	patternDuration float64
}

func NewBoss(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Boss {
	patterns := []interfaces.BulletPattern{
		NewCirclePattern(12, 150),
		NewSpiralPattern(8, 180, 3.0),
		NewWavePattern(10, 160, 300),
		NewSprayPattern(7, 90, 200),
		NewBurstPattern(12, 180, 45),
	}

	return &Boss{
		BaseEntity: types.BaseEntity{
			Position:  position,
			Width:     64, Height: 64,
			Speed:     3,
			Health:    1000,
			MaxHealth: 1000,
		},
		phase:           1,
		eventManager:    eventManager,
		shootCooldown:   NewThreadSafeCooldown(0.5),
		patternSequence: patterns,
		currentPattern:  0,
		patternTimer:    0,
		patternDuration: 3.0,
	}
}

func (b *Boss) Update(deltaTime float64) error {
	b.shootCooldown.Update(deltaTime)
	b.patternTimer += deltaTime

	if b.patternTimer >= b.patternDuration {
		b.nextPattern()
		b.patternTimer = 0
	}

	if b.CanShoot() && b.Shoot() {
	}

	if b.Health <= 500 && b.phase == 1 {
		b.ChangePhase(2)
	}

	return nil
}

func (b *Boss) GetSprite() *graphics.SpriteData {
	level := graphics.Level1
	if b.phase == 2 {
		level = graphics.Level2
	}
	return graphics.GetEnemySprite(graphics.Boss, level)
}

func (b *Boss) nextPattern() {
	b.currentPattern = (b.currentPattern + 1) % len(b.patternSequence)

	if b.phase == 2 {
		b.shootCooldown.SetCooldownTime(0.3)
		b.patternDuration = 2.0
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	// Don't draw dead boss
	if !b.IsAlive() {
		return
	}
	sprite := b.GetSprite()
	if sprite != nil {
		graphics.DrawSprite(screen, sprite, b.Position.X, b.Position.Y)
	}
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
	switch o := other.(type) {
	case *Player:
		return
	case *Bullet:
		if !o.IsEnemyBullet() {
			b.TakeDamage(10)
			b.eventManager.Publish(interfaces.BossDamaged, b)
			if b.Health <= 0 && b.IsAlive() == false {
				// Boss defeated - publish event with position for explosion
				b.eventManager.Publish(interfaces.BossDefeated, map[string]any{
					"boss": b,
					"x":    b.Position.X + b.Width/2,
					"y":    b.Position.Y + b.Height/2,
				})
				b.eventManager.Publish(interfaces.ScoreEvent, 500)
			}
		}
	}
}

func (b *Boss) GetBossType() graphics.EnemyType {
	return graphics.Boss
}

func (b *Boss) CanShoot() bool {
	return b.shootCooldown.CanAct()
}

func (b *Boss) Shoot() bool {
	if b.shootCooldown.TryAct() {
		currentBulletPattern := b.patternSequence[b.currentPattern]

		patternEvent := interfaces.PatternShootEvent{
			Shooter: b,
			Pattern: currentBulletPattern,
		}

		b.eventManager.Publish(interfaces.BossShot, patternEvent)
		return true
	}
	return false
}

func (b *Boss) ChangePhase(newPhase int) {
	b.phase = newPhase
	b.eventManager.Publish(interfaces.BossPhaseChanged, b)

	if newPhase == 2 {
		enhancedPatterns := []interfaces.BulletPattern{
			NewCirclePattern(20, 200),
			NewSprayPattern(9, 150, 250),
			NewBurstPattern(12, 220, 60),
			NewCirclePattern(24, 180),
		}
		b.patternSequence = enhancedPatterns
		b.currentPattern = 0
	}
}

func (b *Boss) GetShootCooldownRemaining() time.Duration {
	return b.shootCooldown.GetRemainingCooldown()
}

func (b *Boss) ResetShootCooldown() {
	b.shootCooldown.Reset()
}

func (b *Boss) SetShootRate(ratePerSecond float64) {
	cooldownTime := 1.0 / ratePerSecond
	b.shootCooldown.SetCooldownTime(cooldownTime)
}

var _ types.GameEntity = (*Boss)(nil)
