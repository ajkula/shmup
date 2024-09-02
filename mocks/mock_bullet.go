package mocks

import (
	"fmt"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/testinterfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type MockBullet struct {
	types.BaseEntity
	isEnemy        bool
	eventManager   interfaces.EventManagerInterface
	outOfBounds    bool
	UpdateCalled   bool
	destroyCalled  bool
	collidedEntity types.Entity
}

func NewMockBullet(x, y float64, isEnemy bool, eventManager interfaces.EventManagerInterface) *MockBullet {
	return &MockBullet{
		BaseEntity: types.BaseEntity{
			Position: types.Vector2D{X: x, Y: y},
			Width:    8, Height: 8,
			Speed:  10,
			Health: 1,
		},
		isEnemy:      isEnemy,
		eventManager: eventManager,
	}
}

func (m *MockBullet) Update(deltaTime float64) error {
	m.UpdateCalled = true
	if m.outOfBounds {
		m.Destroy()
	}
	if m.collidedEntity != nil {
		m.Destroy()
	}
	return nil
}

func (m *MockBullet) Draw(screen *ebiten.Image) {
	// Mock implem
}

func (m *MockBullet) OnCollision(other types.Entity) {
	m.collidedEntity = other
	m.Destroy()
}

func (m *MockBullet) Destroy() {
	if m.IsAlive() {
		m.destroyCalled = true
		m.Health = 0
		fmt.Println(" **************** Will publish BulletDestroyed")
		m.eventManager.Publish(interfaces.BulletDestroyed, m)
		m.Health = -1
	}
}

func (m *MockBullet) CanCollideWith(other types.Entity) bool {
	return true
}

func (m *MockBullet) SetOutOfBounds(outOfBounds bool) {
	m.outOfBounds = outOfBounds
}

func (m *MockBullet) IsEnemyBullet() bool {
	return m.isEnemy
}

var _ testinterfaces.MockBullet = (*MockBullet)(nil)
