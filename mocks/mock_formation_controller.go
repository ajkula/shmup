package mocks

import (
	"sync/atomic"

	"github.com/ajkula/shmup/testinterfaces"
	"github.com/ajkula/shmup/types"
)

type MockFormationController struct {
	updateError     error
	state           int32 // atomic FormationState
	centerPosition  types.Vector2D
	elapsedTime     float64
	enemies         []types.GameEntity
	aliveEnemyCount int32 // atomic
	id              string
}

func NewMockFormationController() *MockFormationController {
	return &MockFormationController{
		state:           int32(types.FormationActive),
		centerPosition:  types.Vector2D{X: 0, Y: 0},
		enemies:         make([]types.GameEntity, 0),
		aliveEnemyCount: 0,
		id:              "mock_formation",
	}
}

// FormationController interface implementation
func (m *MockFormationController) Update(deltaTime float64) error {
	m.elapsedTime += deltaTime
	return m.updateError
}

func (m *MockFormationController) GetState() types.FormationState {
	return types.FormationState(atomic.LoadInt32(&m.state))
}

func (m *MockFormationController) AddEnemy(enemy types.GameEntity) {
	m.enemies = append(m.enemies, enemy)
	atomic.StoreInt32(&m.aliveEnemyCount, int32(len(m.enemies)))
}

func (m *MockFormationController) RemoveEnemy(enemy types.GameEntity) {
	for i, e := range m.enemies {
		if e == enemy {
			m.enemies = append(m.enemies[:i], m.enemies[i+1:]...)
			break
		}
	}
	atomic.StoreInt32(&m.aliveEnemyCount, int32(len(m.enemies)))
}

func (m *MockFormationController) GetEnemies() []types.GameEntity {
	enemies := make([]types.GameEntity, len(m.enemies))
	copy(enemies, m.enemies)
	return enemies
}

func (m *MockFormationController) GetAliveEnemyCount() int {
	return int(atomic.LoadInt32(&m.aliveEnemyCount))
}

func (m *MockFormationController) GetCenterPosition() types.Vector2D {
	return m.centerPosition
}

func (m *MockFormationController) SetCenterPosition(pos types.Vector2D) {
	m.centerPosition = pos
}

func (m *MockFormationController) GetElapsedTime() float64 {
	return m.elapsedTime
}

func (m *MockFormationController) GetID() string {
	return m.id
}

// MockFormationController interface implementation (for testing)
func (m *MockFormationController) SetUpdateError(err error) {
	m.updateError = err
}

func (m *MockFormationController) SetState(state types.FormationState) {
	atomic.StoreInt32(&m.state, int32(state))
}

func (m *MockFormationController) SetAliveEnemyCount(count int) {
	atomic.StoreInt32(&m.aliveEnemyCount, int32(count))
}

// Ensure MockFormationController implements interfaces
var _ types.FormationController = (*MockFormationController)(nil)
var _ testinterfaces.MockFormationController = (*MockFormationController)(nil)
