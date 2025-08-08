package interfaces

import (
	"context"

	"github.com/ajkula/shmup/types"
)

type EventType int

const (
	SystemTick EventType = iota
	CollisionEvent
	InputEvent
	GameStateChangeEvent
	LevelEvent
	LevelChanged
	ScoreEvent
	ScoreReset
	PlayerShot
	PlayerDamaged
	PlayerDestroyed
	BulletCreated
	BulletDestroyed
	EnemyCreated
	EnemyShot
	EnemyDamaged
	EnemyDestroyed
	BossShot
	BossPhaseChanged
	BossDamaged
	BossDefeated
	EnemyAddedToFormation
	EnemyRemovedFromFormation
	FormationCreated
	FormationDestroyed
	FormationStateChanged
	WaveStarted
	WaveCompleted
	EntityMoved
)

type Event struct {
	Type EventType
	Data any
}

type EventManagerInterface interface {
	Initialize(ctx context.Context) error
	Update(deltaTime float64) error
	Run(ctx context.Context) error
	Shutdown()
	Publish(eventType EventType, data any) error
	Subscribe(eventType EventType) (<-chan Event, error)
	Unsubscribe(eventType EventType, ch <-chan Event) error
}

type BulletData struct {
	Position  types.Vector2D
	Direction types.Vector2D
	Speed     float64
}

type BulletPattern interface {
	GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []BulletData
}

type PatternShootEvent struct {
	Shooter types.GameEntity
	Pattern BulletPattern
}
