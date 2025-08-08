package testinterfaces

import (
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type MockEntity interface {
	types.Entity
	SetAlive(bool)
	SetUpdateError(error)
}

type MockBullet interface {
	types.GameEntity
	SetOutOfBounds(bool)
	IsEnemyBullet() bool
}

type MockEnemy interface {
	types.GameEntity
	SetAlive(bool)
	SetUpdateError(error)
}

type MockEventManager interface {
	interfaces.EventManagerInterface
	GetPublishedEvents() []interfaces.Event
	ClearPublishedEvents()
}

type MockFormationController interface {
	types.FormationController
	SetUpdateError(error)
	SetState(types.FormationState)
	SetAliveEnemyCount(int)
}

type MockMovementPattern interface {
	types.MovementPattern
	SetOffsetOverride(types.Vector2D)
	SetCenterMovementOverride(types.Vector2D)
}
