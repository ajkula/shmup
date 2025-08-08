package entity

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

// Formation implements FormationController with thread-safety
type Formation struct {
	mu sync.RWMutex

	// Identity
	id string

	// State management
	state       int32 // atomic FormationState
	elapsedTime float64

	// Position
	centerPosition  types.Vector2D
	initialPosition types.Vector2D

	// Movement
	pattern types.MovementPattern

	// Entities (direct references, not string IDs)
	enemies []types.GameEntity

	// Events
	eventManager interfaces.EventManagerInterface
}

func NewFormation(id string, pattern types.MovementPattern, centerPos types.Vector2D, eventManager interfaces.EventManagerInterface) *Formation {
	return &Formation{
		id:              id,
		state:           int32(types.FormationSpawning),
		centerPosition:  centerPos,
		initialPosition: centerPos,
		pattern:         pattern,
		enemies:         make([]types.GameEntity, 0),
		eventManager:    eventManager,
	}
}

// SOLID: Single Responsibility - Lifecycle management
func (f *Formation) Update(deltaTime float64) error {
	f.mu.Lock()
	f.elapsedTime += deltaTime
	currentTime := f.elapsedTime
	f.mu.Unlock()

	currentState := types.FormationState(atomic.LoadInt32(&f.state))

	switch currentState {
	case types.FormationSpawning:
		f.updateSpawning()
	case types.FormationActive:
		f.updateActive(currentTime)
	case types.FormationExiting:
		f.updateExiting()
	case types.FormationDestroyed:
		// Formation is done, no updates needed
		return nil
	}

	f.checkStateTransitions()
	return nil
}

func (f *Formation) updateSpawning() {
	// All enemies spawned? Move to active
	f.mu.RLock()
	hasEnemies := len(f.enemies) > 0
	f.mu.RUnlock()

	if hasEnemies {
		atomic.StoreInt32(&f.state, int32(types.FormationActive))
		f.eventManager.Publish(interfaces.FormationCreated, f)
	}
}

func (f *Formation) updateActive(currentTime float64) {
	centerMovement := f.pattern.GetCenterMovement(currentTime)
	newCenter := f.initialPosition.Add(centerMovement)

	f.mu.RLock()
	enemies := make([]types.GameEntity, len(f.enemies))
	copy(enemies, f.enemies)
	f.mu.RUnlock()

	positions := make([]types.Vector2D, len(enemies))
	for i, enemy := range enemies {
		if !enemy.IsAlive() {
			positions[i] = enemy.GetPosition()
			continue
		}
		offset := f.pattern.GetOffset(i, currentTime, f)
		positions[i] = newCenter.Add(offset)
	}

	f.mu.Lock()
	f.centerPosition = newCenter
	f.mu.Unlock()

	for i, enemy := range enemies {
		if enemy.IsAlive() {
			enemy.SetPosition(positions[i])
		}
	}
}

func (f *Formation) updateExiting() {
	// Clean up formation
	atomic.StoreInt32(&f.state, int32(types.FormationDestroyed))
	f.eventManager.Publish(interfaces.FormationDestroyed, f)
}

func (f *Formation) checkStateTransitions() {
	// Check if all enemies are dead
	aliveCount := f.GetAliveEnemyCount()
	if aliveCount == 0 && atomic.LoadInt32(&f.state) == int32(types.FormationActive) {
		atomic.StoreInt32(&f.state, int32(types.FormationDestroyed))
		f.eventManager.Publish(interfaces.FormationDestroyed, f)
	}
}

// SOLID: Interface Segregation - State queries
func (f *Formation) GetState() types.FormationState {
	return types.FormationState(atomic.LoadInt32(&f.state))
}

func (f *Formation) GetID() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.id
}

func (f *Formation) GetElapsedTime() float64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.elapsedTime
}

// SOLID: Interface Segregation - Position management
func (f *Formation) GetCenterPosition() types.Vector2D {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.centerPosition
}

func (f *Formation) SetCenterPosition(pos types.Vector2D) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.centerPosition = pos
}

// SOLID: Interface Segregation - Entity management
func (f *Formation) AddEnemy(enemy types.GameEntity) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enemies = append(f.enemies, enemy)
}

func (f *Formation) RemoveEnemy(enemy types.GameEntity) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, e := range f.enemies {
		if e == enemy {
			f.enemies = append(f.enemies[:i], f.enemies[i+1:]...)
			break
		}
	}
}

func (f *Formation) GetEnemies() []types.GameEntity {
	f.mu.RLock()
	defer f.mu.RUnlock()

	enemies := make([]types.GameEntity, len(f.enemies))
	copy(enemies, f.enemies)
	return enemies
}

func (f *Formation) GetAliveEnemyCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	count := 0
	for _, enemy := range f.enemies {
		if enemy.IsAlive() {
			count++
		}
	}
	return count
}

// Debug string representation
func (f *Formation) String() string {
	return fmt.Sprintf("Formation{ID: %s, State: %v, Enemies: %d, Alive: %d}",
		f.GetID(), f.GetState(), len(f.enemies), f.GetAliveEnemyCount())
}

// Ensure Formation implements FormationController
var _ types.FormationController = (*Formation)(nil)
