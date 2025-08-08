package types

// SOLID: Single Responsibility Principle
// Each interface has ONE specific responsibility

// FormationState represents the lifecycle of a formation
type FormationState int

const (
	FormationSpawning  FormationState = iota // Enemies are being spawned
	FormationActive                          // Formation is moving and active
	FormationExiting                         // Formation is leaving the screen
	FormationDestroyed                       // All enemies destroyed or exited
)

// MovementPattern defines how formations move (Strategy Pattern)
type MovementPattern interface {
	// Calculate position offset for enemy at index i at time t
	GetOffset(enemyIndex int, time float64, formation FormationController) Vector2D
	// Get the formation's center movement
	GetCenterMovement(time float64) Vector2D
}

// FormationController manages formation lifecycle and state
type FormationController interface {
	// Lifecycle
	Update(deltaTime float64) error
	GetState() FormationState

	// Entity management
	AddEnemy(enemy GameEntity)
	RemoveEnemy(enemy GameEntity)
	GetEnemies() []GameEntity
	GetAliveEnemyCount() int

	// Position and movement
	GetCenterPosition() Vector2D
	SetCenterPosition(pos Vector2D)

	// Timing
	GetElapsedTime() float64

	// Identification
	GetID() string
}

// FormationSpawner creates formations (Factory Pattern)
type FormationSpawner interface {
	SpawnFormation(pattern MovementPattern, centerPos Vector2D, enemyCount int) FormationController
}

// FormationRenderer handles formation rendering (if needed for debug)
type FormationRenderer interface {
	RenderFormationDebug(formation FormationController, screen interface{})
}
