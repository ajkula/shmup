package types

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type BaseEntity struct {
	Position      Vector2D
	Width, Height float64
	Speed         float64
	Health        int
	Color         color.Color
}

func (e *BaseEntity) GetPosition() Vector2D {
	return e.Position
}

func (e *BaseEntity) SetPosition(pos Vector2D) {
	e.Position = pos
}

func (e *BaseEntity) GetSize() (float64, float64) {
	return e.Width, e.Height
}

func (e *BaseEntity) GetHealth() int {
	return e.Health
}

func (e *BaseEntity) TakeDamage(amount int) {
	e.Health -= amount
	if e.Health < 0 {
		e.Health = 0
	}
}

func (e *BaseEntity) IsAlive() bool {
	return e.Health > 0
}

func (e *BaseEntity) GetCollisionBox() (x, y, width, height float64) {
	return e.Position.X, e.Position.Y, e.Width, e.Height
}

func (e *BaseEntity) GetColor() color.Color {
	return e.Color
}

func (e *BaseEntity) Update(deltaTime float64) error {
	return nil
}

func (e *BaseEntity) Draw(screen *ebiten.Image) {
	// default
}

var _ Entity = (*BaseEntity)(nil)
