package types

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// CORE TYPES - Keep these, they're actively used

type Vector2D struct {
	X, Y float64
}

func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{v.X + other.X, v.Y + other.Y}
}

func (v Vector2D) Subtract(other Vector2D) Vector2D {
	return Vector2D{v.X - other.X, v.Y - other.Y}
}

func (v Vector2D) Multiply(scalar float64) Vector2D {
	return Vector2D{v.X * scalar, v.Y * scalar}
}

// CORE INTERFACES - Keep these, they're the foundation

// Base entity interface - used by all game objects
type Entity interface {
	Update(deltaTime float64) error
	Draw(screen *ebiten.Image)
	GetPosition() Vector2D
	SetPosition(pos Vector2D)
	GetSize() (width, height float64)
	GetCollisionBox() (x, y, width, height float64)
	IsAlive() bool
	TakeDamage(amount int)
	GetHealth() int
	GetColor() color.Color
}

// Can be updated
type Updatable interface {
	Update(deltaTime float64) error
}

// Can be drawn to screen
type Renderable interface {
	Draw(screen *ebiten.Image)
}

// Can participate in collisions
type Collidable interface {
	CanCollideWith(other Entity) bool
	OnCollision(other Entity)
}

// Combination of Entity and Collidable - used by Player, Enemy, Bullet, Boss
type GameEntity interface {
	Entity
	Collidable
}

// Event handling
type EventHandler interface {
	HandleEvent(event interface{}) error
}
