package mocks

import (
	"image/color"

	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type MockEnemy struct {
	UpdateCalled bool
	Alive        bool
	UpdateError  error
	Health       int
	Position     types.Vector2D
	eventManager interfaces.EventManagerInterface
}

func NewMockEnemy(eventManager interfaces.EventManagerInterface) *MockEnemy {
	return &MockEnemy{
		Alive:        true,
		Health:       20,
		eventManager: eventManager,
	}
}

func (m *MockEnemy) Update(deltaTime float64) error {
	m.UpdateCalled = true
	return m.UpdateError
}

func (m *MockEnemy) OnCollision(other types.Entity) {
	m.TakeDamage(10)
	if m.Health <= 0 {
		m.Alive = false
	}
}

func (m *MockEnemy) Draw(screen *ebiten.Image)              {}
func (m *MockEnemy) GetPosition() types.Vector2D            { return m.Position }
func (m *MockEnemy) SetPosition(pos types.Vector2D)         { m.Position = pos }
func (m *MockEnemy) GetSize() (width, height float64)       { return 0, 0 }
func (m *MockEnemy) GetCollisionBox() (x, y, w, h float64)  { return 0, 0, 0, 0 }
func (m *MockEnemy) IsAlive() bool                          { return m.Alive }
func (m *MockEnemy) TakeDamage(amount int)                  {}
func (m *MockEnemy) GetHealth() int                         { return 0 }
func (m *MockEnemy) GetColor() color.Color                  { return color.White }
func (m *MockEnemy) CanCollideWith(other types.Entity) bool { return false }
func (m *MockEnemy) GetSprite() *graphics.SpriteData        { return nil }
