package interfaces

import "context"

type EventType int

const (
	CollisionEvent EventType = iota
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
