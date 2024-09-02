package entity

import (
	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Bullet struct {
	types.BaseEntity
	isEnemy      bool
	eventManager interfaces.EventManagerInterface
	direction    types.Vector2D
}

func NewBullet(x, y float64, isEnemy bool, eventManager interfaces.EventManagerInterface) *Bullet {
	direction := types.Vector2D{X: 0, Y: -1}
	if isEnemy {
		direction.Y = 1
	}

	return &Bullet{
		BaseEntity: types.BaseEntity{
			Position: types.Vector2D{X: x, Y: y},
			Width:    8, Height: 8,
			Speed:  10,
			Health: 1,
		},
		isEnemy:      isEnemy,
		eventManager: eventManager,
		direction:    direction,
	}
}

func (b *Bullet) Update(deltaTime float64) error {
	newPos := b.GetPosition()
	newPos = newPos.Add(b.direction.Multiply(b.Speed * deltaTime))
	b.SetPosition(newPos)

	if b.IsOutOfBounds() {
		b.Destroy()
	}

	return nil
}

func (b *Bullet) Destroy() {
	if b.IsAlive() {
		b.Health = 0
		b.eventManager.Publish(interfaces.BulletDestroyed, b)
	}
}

func (b *Bullet) IsOutOfBounds() bool {
	pos := b.GetPosition()
	return pos.X < 0 || pos.X > float64(config.Config.ScreenWidth) || pos.Y < 0 || pos.Y > float64(config.Config.ScreenHeight)
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	// TODO
}

func (b *Bullet) CanCollideWith(other types.Entity) bool {
	if b.isEnemy {
		_, isPlayer := other.(types.GameEntity)
		return isPlayer
	}
	_, isEnemy := other.(types.GameEntity)
	return isEnemy
}

func (b *Bullet) OnCollision(other types.Entity) {
	b.Destroy()
}

func (b *Bullet) IsEnemyBullet() bool {
	return b.isEnemy
}

var _ types.GameEntity = (*Bullet)(nil)
