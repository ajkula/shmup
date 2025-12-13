// Nouveau fichier : system/explosion_system.go

package system

import (
	"context"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/effects"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type ExplosionSystem struct {
	core.BaseSystem
	explosionManager *effects.ExplosionManager
	eventManager     interfaces.EventManagerInterface
	eventChannels    map[interfaces.EventType]<-chan interfaces.Event
	isShutdown       int32
}

func NewExplosionSystem(explosionManager *effects.ExplosionManager, eventManager interfaces.EventManagerInterface) *ExplosionSystem {
	return &ExplosionSystem{
		explosionManager: explosionManager,
		eventManager:     eventManager,
		eventChannels:    make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (es *ExplosionSystem) Initialize(ctx context.Context) error {
	if err := es.BaseSystem.Initialize(ctx); err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.EnemyDestroyed,
		interfaces.BossDefeated,
		interfaces.PlayerDestroyed,
	}

	for _, eventType := range eventTypes {
		ch, err := es.eventManager.Subscribe(eventType)
		if err != nil {
			return err
		}
		es.eventChannels[eventType] = ch
	}

	go es.eventListener()
	return nil
}

func (es *ExplosionSystem) eventListener() {
	for {
		select {
		case <-es.CTX.Done():
			return
		case evt := <-es.eventChannels[interfaces.EnemyDestroyed]:
			es.handleDestruction(evt)
		case evt := <-es.eventChannels[interfaces.BossDefeated]:
			es.handleDestruction(evt)
		case evt := <-es.eventChannels[interfaces.PlayerDestroyed]:
			es.handleDestruction(evt)
		}
	}
}

func (es *ExplosionSystem) handleDestruction(evt interfaces.Event) {
	// Try direct entity first
	if entity, ok := evt.Data.(types.Entity); ok && entity != nil {
		es.explosionManager.CreateExplosionFromEntity(entity)
		return
	}

	// Handle boss defeated event which sends a map
	if dataMap, ok := evt.Data.(map[string]any); ok {
		if boss, ok := dataMap["boss"].(types.Entity); ok && boss != nil {
			es.explosionManager.CreateExplosionFromEntity(boss)
		}
	}
}

func (es *ExplosionSystem) Update(deltaTime float64) error {
	select {
	case <-es.CTX.Done():
		return es.CTX.Err()
	default:
		return nil
	}
}

func (es *ExplosionSystem) Shutdown() {
	if !atomic.CompareAndSwapInt32(&es.isShutdown, 0, 1) {
		return
	}

	for eventType, ch := range es.eventChannels {
		es.eventManager.Unsubscribe(eventType, ch)
	}
	es.eventChannels = nil
}

var _ core.System = (*ExplosionSystem)(nil)
