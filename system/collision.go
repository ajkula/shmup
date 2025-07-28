package system

import (
	"context"
	"sync"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	// Add this line to import the package that contains the definition of interfaces.EntityMoved
)

const (
	fixedDeltaTime = 1.0 / 60.0
)

type CollisionSystem struct {
	core.BaseSystem
	quadtree      *Quadtree
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	mu            sync.RWMutex
}

type CollisionData struct {
	EntityA types.GameEntity
	EntityB types.GameEntity
}

func NewCollisionSystem(eventManager interfaces.EventManagerInterface) *CollisionSystem {
	worldBounds := Rect{0, 0, float64(config.Config.ScreenWidth), float64(config.Config.ScreenHeight)}
	return &CollisionSystem{
		quadtree:     NewQuadtree(worldBounds, 4),
		eventManager: eventManager,
	}
}

func (cs *CollisionSystem) Initialize(ctx context.Context) error {
	if err := cs.BaseSystem.Initialize(ctx); err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.BulletCreated,
		interfaces.BulletDestroyed,
		interfaces.EnemyDestroyed,
		interfaces.BossDefeated,
		interfaces.PlayerDestroyed,
		interfaces.EntityMoved,
	}

	cs.eventChannels = make(map[interfaces.EventType]<-chan interfaces.Event)
	for _, eventType := range eventTypes {
		ch, err := cs.eventManager.Subscribe(eventType)
		if err != nil {
			return err
		}
		cs.eventChannels[eventType] = ch
	}

	return nil
}

func (cs *CollisionSystem) Update(deltaTime float64) error {
	select {
	case <-cs.CTX.Done():
		return cs.CTX.Err()
	default:
		cs.processEvents()
		cs.checkCollisions()
	}
	return nil
}

func (cs *CollisionSystem) processEvents() {
	for eventType, ch := range cs.eventChannels {
		select {
		case evt, ok := <-ch:
			if !ok {
				cs.eventManager.Unsubscribe(eventType, ch)
				continue
			}
			cs.handleEvent(eventType, evt)
		default:
			// No events for this type
		}
	}
}

func (cs *CollisionSystem) handleEvent(eventType interfaces.EventType, evt interfaces.Event) {
	entity, ok := evt.Data.(types.GameEntity)
	if !ok {
		return
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	switch eventType {
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

				// Publier un événement de collision
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
	cs.mu.Lock()
	defer cs.mu.Unlock()
	// Clear the quadtree and close any channels if necessary
	cs.quadtree = nil
}
