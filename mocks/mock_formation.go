package mocks

import (
	"image/color"

	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type MockFormation struct {
	UpdateCalled bool
	updateError  error
}

func (m *MockFormation) Update(deltaTime float64) error {
	m.UpdateCalled = true
	return m.updateError
}

func (m *MockFormation) Draw(screen *ebiten.Image)                {}
func (m *MockFormation) GetPosition() types.Vector2D              { return types.Vector2D{} }
func (m *MockFormation) SetPosition(pos types.Vector2D)           {}
func (m *MockFormation) GetSize() (width, height float64)         { return 0, 0 }
func (m *MockFormation) GetCollisionBox() (x, y, w, h float64)    { return 0, 0, 0, 0 }
func (m *MockFormation) IsAlive() bool                            { return true }
func (m *MockFormation) TakeDamage(amount int)                    {}
func (m *MockFormation) GetHealth() int                           { return 0 }
func (m *MockFormation) GetColor() color.Color                    { return color.White }
func (m *MockFormation) GetEntities() []types.Entity              { return nil }
func (m *MockFormation) AddEntity(e types.Entity)                 {}
func (m *MockFormation) RemoveEntity(e types.Entity)              {}
func (m *MockFormation) IsComplete() bool                         { return false }
func (m *MockFormation) GetFormationType() types.FormationType    { return types.LineFormation }
func (m *MockFormation) SetPattern(pattern types.MovementPattern) {}
