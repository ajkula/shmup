package types

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type FormationMember interface {
	GetPosition() Vector2D
	SetPosition(pos Vector2D)
}

type Vector2D struct {
	X, Y float64
}

// entité de base dans le jeu
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

// peut être mise à jour
type Updatable interface {
	Update(deltaTime float64) error
}

// peut être dessinée
type Renderable interface {
	Draw(screen *ebiten.Image)
}

type Collidable interface {
	CanCollideWith(other Entity) bool
	OnCollision(other Entity)
}

// Entity / Collidable
type GameEntity interface {
	Entity
	Collidable
}

// mouvement pour les formations
type MovementPattern interface {
	Move(members []FormationMember, elapsedTime float64)
}

// groupe d'ennemis
type Formation interface {
	Entity
	GetEntities() []Entity
	AddEntity(e Entity)
	RemoveEntity(e Entity)
	IsComplete() bool
	GetFormationType() FormationType
	SetPattern(pattern MovementPattern)
}

// types de formations
type FormationType int

const (
	LineFormation FormationType = iota
	ColumnFormation
	VFormation
	CircleFormation
)

// game events def
type EventHandler interface {
	HandleEvent(event interface{}) error
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
