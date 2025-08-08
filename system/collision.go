package system

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type CollisionSystem struct {
	core.BaseSystem
	quadtree      *Quadtree
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	mu            sync.RWMutex
	isShutdown    int32
}

type CollisionData struct {
	EntityA types.GameEntity
	EntityB types.GameEntity
}

func NewCollisionSystem(eventManager interfaces.EventManagerInterface) *CollisionSystem {
	worldBounds := Rect{0, 0, float64(config.Config.ScreenWidth), float64(config.Config.ScreenHeight)}
	return &CollisionSystem{
		quadtree:      NewQuadtree(worldBounds, 4),
		eventManager:  eventManager,
		eventChannels: make(map[interfaces.EventType]<-chan interfaces.Event),
	}
}

func (cs *CollisionSystem) Initialize(ctx context.Context) error {
	if err := cs.BaseSystem.Initialize(ctx); err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.SystemTick,
		interfaces.BulletCreated,
		interfaces.BulletDestroyed,
		interfaces.EnemyDestroyed,
		interfaces.BossDefeated,
		interfaces.PlayerDestroyed,
		interfaces.EntityMoved,
	}

	for _, eventType := range eventTypes {
		ch, err := cs.eventManager.Subscribe(eventType)
		if err != nil {
			return err
		}
		cs.eventChannels[eventType] = ch
	}

	go cs.eventListener()

	return nil
}

func (cs *CollisionSystem) Update(deltaTime float64) error {
	select {
	case <-cs.CTX.Done():
		return cs.CTX.Err()
	default:
		return nil
	}
}

func (cs *CollisionSystem) eventListener() {
	systemTickCh := cs.eventChannels[interfaces.SystemTick]
	bulletCreatedCh := cs.eventChannels[interfaces.BulletCreated]
	bulletDestroyedCh := cs.eventChannels[interfaces.BulletDestroyed]
	enemyDestroyedCh := cs.eventChannels[interfaces.EnemyDestroyed]
	bossDefeatedCh := cs.eventChannels[interfaces.BossDefeated]
	playerDestroyedCh := cs.eventChannels[interfaces.PlayerDestroyed]
	entityMovedCh := cs.eventChannels[interfaces.EntityMoved]

	for {
		select {
		case <-cs.CTX.Done():
			return
		case _, ok := <-systemTickCh:
			if !ok {
				return
			}
			cs.processAllAvailableEvents()
			cs.checkCollisions()
		case evt, ok := <-bulletCreatedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		case evt, ok := <-bulletDestroyedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		case evt, ok := <-enemyDestroyedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		case evt, ok := <-bossDefeatedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		case evt, ok := <-playerDestroyedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		case evt, ok := <-entityMovedCh:
			if !ok {
				return
			}
			cs.handleEntityEvent(evt)
		}
	}
}

func (cs *CollisionSystem) processAllAvailableEvents() {
	if atomic.LoadInt32(&cs.isShutdown) == 1 {
		return
	}

	eventChannels := []<-chan interfaces.Event{
		cs.eventChannels[interfaces.BulletCreated],
		cs.eventChannels[interfaces.BulletDestroyed],
		cs.eventChannels[interfaces.EnemyDestroyed],
		cs.eventChannels[interfaces.BossDefeated],
		cs.eventChannels[interfaces.PlayerDestroyed],
		cs.eventChannels[interfaces.EntityMoved],
	}

	for _, ch := range eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				cs.handleEntityEvent(evt)
			default:
				goto nextChannel
			}
		}
	nextChannel:
	}
}

func (cs *CollisionSystem) handleEntityEvent(evt interfaces.Event) {
	entity, ok := evt.Data.(types.GameEntity)
	if !ok {
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	switch evt.Type {
	case interfaces.BulletCreated, interfaces.EntityMoved:
		cs.quadtree.Insert(entity)
	case interfaces.BulletDestroyed, interfaces.EnemyDestroyed, interfaces.BossDefeated, interfaces.PlayerDestroyed:
		cs.quadtree.Remove(entity)
	}
}

func (cs *CollisionSystem) checkCollisions() {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entities := cs.quadtree.GetAllEntities()
	for _, entity := range entities {
		bounds := cs.getEntityBounds(entity)
		potentialCollisions := cs.quadtree.Query(bounds)

		for _, other := range potentialCollisions {
			if entity == other {
				continue
			}
			if entity.CanCollideWith(other) && cs.detectCollision(entity, other) {
				entity.OnCollision(other)
				other.OnCollision(entity)

				cs.eventManager.Publish(interfaces.CollisionEvent, CollisionData{EntityA: entity, EntityB: other})
			}
		}
	}
}

func (cs *CollisionSystem) getEntityBounds(entity types.GameEntity) Rect {
	x, y, w, h := entity.GetCollisionBox()
	return Rect{x, y, w, h}
}

func (cs *CollisionSystem) detectCollision(a, b types.GameEntity) bool {
	ax, ay, aw, ah := a.GetCollisionBox()
	bx, by, bw, bh := b.GetCollisionBox()

	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func (cs *CollisionSystem) Shutdown() {
	if !atomic.CompareAndSwapInt32(&cs.isShutdown, 0, 1) {
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	for eventType, ch := range cs.eventChannels {
		cs.eventManager.Unsubscribe(eventType, ch)
	}
	cs.eventChannels = nil
	cs.quadtree = nil
}
