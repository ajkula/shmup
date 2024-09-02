package manager

import (
	"context"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type EnemyManager struct {
	core.BaseSystem
	enemies       []types.GameEntity
	formations    []types.Formation
	eventManager  interfaces.EventManagerInterface
	mu            sync.RWMutex
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
}

func NewEnemyManager(eventManager interfaces.EventManagerInterface) *EnemyManager {
	return &EnemyManager{
		enemies:       make([]types.GameEntity, 0),
		formations:    make([]types.Formation, 0),
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
		interfaces.FormationCreated,
		interfaces.FormationDestroyed,
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
		em.processEvents()
		return em.updateEntities(deltaTime)
	}
}

func (em *EnemyManager) processEvents() {
	for eventType, ch := range em.eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				em.handleEvent(eventType, evt)
			default:
				// No more events for this type
				return
			}
		}
	}
}

func (em *EnemyManager) handleEvent(eventType interfaces.EventType, evt interfaces.Event) {
	switch eventType {
	case interfaces.EnemyCreated:
		if enemy, ok := evt.Data.(types.GameEntity); ok {
			em.AddEnemy(enemy)
		}
	case interfaces.EnemyDestroyed:
		if enemy, ok := evt.Data.(types.GameEntity); ok {
			em.RemoveEnemy(enemy)
		}
	case interfaces.FormationCreated:
		if formation, ok := evt.Data.(types.Formation); ok {
			em.AddFormation(formation)
		}
	case interfaces.FormationDestroyed:
		if formation, ok := evt.Data.(types.Formation); ok {
			em.RemoveFormation(formation)
		}
	}
}

func (em *EnemyManager) updateEntities(deltaTime float64) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	aliveEnemies := make([]types.GameEntity, 0, len(em.enemies))
	for _, enemy := range em.enemies {
		if err := enemy.Update(deltaTime); err != nil {
			return err
		}
		if enemy.IsAlive() {
			aliveEnemies = append(aliveEnemies, enemy)
		}
	}
	em.enemies = aliveEnemies

	for _, formation := range em.formations {
		if err := formation.Update(deltaTime); err != nil {
			return err
		}
	}
	return nil
}

func (em *EnemyManager) Draw(screen *ebiten.Image) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	for _, enemy := range em.enemies {
		enemy.Draw(screen)
	}
}

func (em *EnemyManager) AddEnemy(enemy types.GameEntity) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.enemies = append(em.enemies, enemy)
}

func (em *EnemyManager) RemoveEnemy(enemy types.GameEntity) {
	em.mu.Lock()
	defer em.mu.Unlock()
	for i, e := range em.enemies {
		if e == enemy {
			em.enemies = append(em.enemies[:i], em.enemies[i+1:]...)
			break
		}
	}
}

func (em *EnemyManager) AddFormation(formation types.Formation) {
	em.mu.Lock()
	defer em.mu.Unlock()
	em.formations = append(em.formations, formation)
}

func (em *EnemyManager) RemoveFormation(formation types.Formation) {
	em.mu.Lock()
	defer em.mu.Unlock()
	for i, f := range em.formations {
		if f == formation {
			em.formations = append(em.formations[:i], em.formations[i+1:]...)
			break
		}
	}
}

func (em *EnemyManager) Shutdown() {
	em.mu.Lock()
	defer em.mu.Unlock()
	for eventType, ch := range em.eventChannels {
		em.eventManager.Unsubscribe(eventType, ch)
	}
	em.eventChannels = nil
	em.enemies = nil
	em.formations = nil
}

var _ core.System = (*EnemyManager)(nil)
