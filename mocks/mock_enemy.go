package mocks

import (
	"image/color"

	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type MockEnemy struct {
	UpdateCalled bool
	Alive        bool
	UpdateError  error
}

func (m *MockEnemy) Update(deltaTime float64) error {
	m.UpdateCalled = true
	return m.UpdateError
}

func (m *MockEnemy) Draw(screen *ebiten.Image)              {}
func (m *MockEnemy) GetPosition() types.Vector2D            { return types.Vector2D{} }
func (m *MockEnemy) SetPosition(pos types.Vector2D)         {}
func (m *MockEnemy) GetSize() (width, height float64)       { return 0, 0 }
func (m *MockEnemy) GetCollisionBox() (x, y, w, h float64)  { return 0, 0, 0, 0 }
func (m *MockEnemy) IsAlive() bool                          { return m.Alive }
func (m *MockEnemy) TakeDamage(amount int)                  {}
func (m *MockEnemy) GetHealth() int                         { return 0 }
func (m *MockEnemy) GetColor() color.Color                  { return color.White }
func (m *MockEnemy) CanCollideWith(other types.Entity) bool { return false }
func (m *MockEnemy) OnCollision(other types.Entity)         {}
