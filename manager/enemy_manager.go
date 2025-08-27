package manager

import (
	"context"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/registry"
	"github.com/ajkula/shmup/types"
)

type EnemyManager struct {
	core.BaseSystem
	enemies       []types.GameEntity
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	isShutdown    int32
}

func NewEnemyManager(eventManager interfaces.EventManagerInterface) *EnemyManager {
	return &EnemyManager{
		enemies:       make([]types.GameEntity, 0),
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (em *EnemyManager) Initialize(ctx context.Context) error {
	err := em.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.EnemyCreated,
		interfaces.EnemyDestroyed,
	}

	for _, eventType := range eventTypes {
		ch, err := em.eventManager.Subscribe(eventType)
		if err != nil {
			return err
		}
		em.eventChannels[eventType] = ch
	}

	return nil
}

func (em *EnemyManager) Update(deltaTime float64) error {
	select {
	case <-em.CTX.Done():
		return em.CTX.Err()
	default:
		em.processAllEvents()

		em.updateEnemies()

		return nil
	}
}

func (em *EnemyManager) processAllEvents() {
	if atomic.LoadInt32(&em.isShutdown) == 1 {
		return
	}

	createdCh := em.eventChannels[interfaces.EnemyCreated]
	for {
		select {
		case evt, ok := <-createdCh:
			if !ok {
				return
			}
			if enemy, ok := evt.Data.(types.GameEntity); ok {
				em.enemies = append(em.enemies, enemy)
			}
		default:
			goto processDestroyed
		}
	}

processDestroyed:
	destroyedCh := em.eventChannels[interfaces.EnemyDestroyed]
	for {
		select {
		case evt, ok := <-destroyedCh:
			if !ok {
				return
			}
			if enemy, ok := evt.Data.(types.GameEntity); ok {
				for i, e := range em.enemies {
					if e == enemy {
						em.enemies = append(em.enemies[:i], em.enemies[i+1:]...)
						break
					}
				}
			}
		default:
			return
		}
	}
}

func (em *EnemyManager) updateEnemies() {
	aliveEnemies := make([]types.GameEntity, 0, len(em.enemies))
	for _, enemy := range em.enemies {
		if err := enemy.Update(core.FixedDeltaTime); err != nil {
			continue
		}
		if enemy.IsAlive() {
			aliveEnemies = append(aliveEnemies, enemy)
		}
	}
	em.enemies = aliveEnemies
}

func (em *EnemyManager) GetRenderableEntities() []types.Renderable {
	renderables := make([]types.Renderable, 0, len(em.enemies))
	for _, enemy := range em.enemies {
		if enemy.IsAlive() {
			renderables = append(renderables, enemy)
		}
	}
	return renderables
}

func (em *EnemyManager) GetEnemyCount() int {
	return len(em.enemies)
}

func (em *EnemyManager) GetEnemies() []types.GameEntity {
	enemies := make([]types.GameEntity, len(em.enemies))
	copy(enemies, em.enemies)
	return enemies
}

func (em *EnemyManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&em.isShutdown, 0, 1) {
		return
	}

	for eventType, ch := range em.eventChannels {
		em.eventManager.Unsubscribe(eventType, ch)
	}
	em.eventChannels = nil
	em.enemies = nil
}

var _ registry.EntityProvider = (*EnemyManager)(nil)
var _ core.System = (*EnemyManager)(nil)
