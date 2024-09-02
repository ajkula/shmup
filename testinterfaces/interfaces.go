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

type MockFormation interface {
	types.Formation
	SetUpdateCalled(bool)
}

type MockEventManager interface {
	interfaces.EventManagerInterface
	GetPublishedEvents() []interfaces.Event
}
