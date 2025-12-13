package entity

import (
	"image/color"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
			Speed:  300,
			Health: 1,
		},
		isEnemy:      isEnemy,
		eventManager: eventManager,
		direction:    direction,
	}
}

func NewBulletWithDirection(x, y float64, direction types.Vector2D, speed float64, isEnemy bool, eventManager interfaces.EventManagerInterface) *Bullet {
	return &Bullet{
		BaseEntity: types.BaseEntity{
			Position: types.Vector2D{X: x, Y: y},
			Width:    8, Height: 8,
			Speed:  speed,
			Health: 1,
		},
		isEnemy:      isEnemy,
		eventManager: eventManager,
		direction:    direction,
	}
}

func (b *Bullet) Update(deltaTime float64) error {
	oldPos := b.GetPosition()
	newPos := oldPos.Add(b.direction.Multiply(b.Speed * deltaTime))
	b.SetPosition(newPos)

	// if b.isEnemy && oldPos.Y < 100 { // TEMP LOG
	// 	fmt.Printf("🔴 Enemy bullet move: (%.1f, %.1f) → (%.1f, %.1f)\n", oldPos.X, oldPos.Y, newPos.X, newPos.Y)
	// }

	if b.IsOutOfBounds() {
		// if b.isEnemy { // TEMP LOG
		// 	fmt.Printf("💀 Enemy bullet OOB: (%.1f, %.1f) bounds: 0-%.1f, 0-%.1f\n",
		// 		newPos.X, newPos.Y, float64(config.Config.ScreenWidth), float64(config.Config.ScreenHeight))
		// }
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

func (b *Bullet) GetDamage() int {
	if b.isEnemy {
		return 10
	}
	return 20
}

func (b *Bullet) IsOutOfBounds() bool {
	pos := b.GetPosition()
	return pos.X < -10 || pos.X > float64(config.Config.ScreenWidth+10) ||
		pos.Y < -10 || pos.Y > float64(config.Config.ScreenHeight+10)
}

func (b *Bullet) Draw(screen *ebiten.Image) {
	pos := b.GetPosition()

	if b.isEnemy {
		bulletColor := color.RGBA{255, 0, 0, 255}
		width := float32(b.Width * 0.6)
		height := float32(b.Height * 1.5)

		x := float32(pos.X) + (float32(b.Width)-width)/2
		y := float32(pos.Y) + (float32(b.Height)-height)/2

		vector.DrawFilledRect(screen, x, y, width, height, bulletColor, false)

	} else {
		bulletColor := color.RGBA{255, 255, 0, 255}

		centerX := float32(pos.X) + float32(b.Width)/2
		centerY := float32(pos.Y) + float32(b.Height)/2
		size := float32(b.Width) * 0.7

		vector.DrawFilledRect(screen,
			centerX-1, centerY-size/2,
			2, size, bulletColor, false)
		vector.DrawFilledRect(screen,
			centerX-size/2, centerY-1,
			size, 2, bulletColor, false)
		vector.DrawFilledRect(screen,
			centerX-2, centerY-2,
			4, 4, bulletColor, false)
	}
}

func (b *Bullet) CanCollideWith(other types.Entity) bool {
	switch other.(type) {
	case *Bullet:
		return false
	case *Player:
		return b.isEnemy
	case *Enemy, *Boss:
		return !b.isEnemy
	default:
		return false
	}
}

func (b *Bullet) OnCollision(other types.Entity) {
	b.Destroy()
}

func (b *Bullet) GetSprite() *graphics.SpriteData {
	return nil
}

func (b *Bullet) IsEnemyBullet() bool {
	return b.isEnemy
}

var _ types.GameEntity = (*Bullet)(nil)
