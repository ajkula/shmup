package system

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type EntityProvider interface {
	GetRenderableEntities() []types.Renderable
}

type CollisionSystem struct {
	core.BaseSystem
	quadtree       *Quadtree
	player         types.GameEntity
	entityRegistry EntityProvider
	eventManager   interfaces.EventManagerInterface
	eventChannels  map[interfaces.EventType]<-chan interfaces.Event
	mu             sync.RWMutex
	isShutdown     int32
	processedPairs map[collisionPair]bool
}

type collisionPair struct {
	id1, id2 string
}

type CollisionData struct {
	EntityA types.GameEntity
	EntityB types.GameEntity
}

func NewCollisionSystem(eventManager interfaces.EventManagerInterface) *CollisionSystem {
	worldBounds := Rect{0, 0, float64(config.Config.ScreenWidth), float64(config.Config.ScreenHeight)}
	return &CollisionSystem{
		quadtree:       NewQuadtree(worldBounds, 4),
		eventManager:   eventManager,
		eventChannels:  make(map[interfaces.EventType]<-chan interfaces.Event),
		processedPairs: make(map[collisionPair]bool),
	}
}

func (cs *CollisionSystem) SetEntityRegistry(registry EntityProvider) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.entityRegistry = registry
}

func (cs *CollisionSystem) SetPlayer(player types.GameEntity) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.player = player
}

func (cs *CollisionSystem) Initialize(ctx context.Context) error {
	if err := cs.BaseSystem.Initialize(ctx); err != nil {
		return err
	}

	// rebuild quadtree at each frame
	eventTypes := []interfaces.EventType{
		interfaces.SystemTick,
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

	for {
		select {
		case <-cs.CTX.Done():
			return
		case _, ok := <-systemTickCh:
			if !ok {
				return
			}
			cs.checkCollisions()
		}
	}
}

func (cs *CollisionSystem) checkCollisions() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Clear completely
	worldBounds := Rect{0, 0, float64(config.Config.ScreenWidth), float64(config.Config.ScreenHeight)}
	cs.quadtree = NewQuadtree(worldBounds, 4)

	// Clear processed pairs
	cs.processedPairs = make(map[collisionPair]bool)

	// Get all entities from the registry
	var allEntities []types.GameEntity

	// Add player if alive
	if cs.player != nil && cs.player.IsAlive() {
		allEntities = append(allEntities, cs.player)
		cs.quadtree.Insert(cs.player)
	}

	// Get entities from registry if available
	if cs.entityRegistry != nil {
		renderables := cs.entityRegistry.GetRenderableEntities()
		for _, r := range renderables {
			if entity, ok := r.(types.GameEntity); ok && entity.IsAlive() {
				allEntities = append(allEntities, entity)
				cs.quadtree.Insert(entity)
			}
		}
	}

	// Debug logging
	var frameCounter int
	frameCounter++
	if frameCounter%60 == 0 {
		var playerCount, enemyCount, bulletCount, bossCount int
		for _, e := range allEntities {
			switch e.(type) {
			case interface{ GetPlayerClass() interface{} }:
				playerCount++
			case interface{ GetEnemyType() interface{} }:
				if _, isBoss := e.(interface{ GetBossType() interface{} }); isBoss {
					bossCount++
				} else {
					enemyCount++
				}
			case interface{ IsEnemyBullet() bool }:
				bulletCount++
			}
		}
		fmt.Printf("Collision: Player=%d, Enemies=%d, Boss=%d, Bullets=%d, Total=%d\n",
			playerCount, enemyCount, bossCount, bulletCount, len(allEntities))
	}

	// Check collisions between all entities
	for _, entity := range allEntities {
		if !entity.IsAlive() {
			continue
		}

		bounds := cs.getEntityBounds(entity)
		potentialCollisions := cs.quadtree.Query(bounds)

		for _, other := range potentialCollisions {
			if entity == other || !other.IsAlive() {
				continue
			}

			// Create collision pair
			pair := cs.makeCollisionPair(entity, other)

			// Skip if already processed
			if cs.processedPairs[pair] {
				continue
			}

			// Check if entities can collide
			if entity.CanCollideWith(other) && cs.detectCollision(entity, other) {
				// Mark as processed
				cs.processedPairs[pair] = true

				// Process collision for both entities
				entity.OnCollision(other)
				other.OnCollision(entity)

				// Publish collision event
				cs.eventManager.Publish(interfaces.CollisionEvent, CollisionData{
					EntityA: entity,
					EntityB: other,
				})
			}
		}
	}
}

func (cs *CollisionSystem) makeCollisionPair(a, b types.GameEntity) collisionPair {
	id1 := fmt.Sprintf("%p", a)
	id2 := fmt.Sprintf("%p", b)

	if id1 < id2 {
		return collisionPair{id1: id1, id2: id2}
	}
	return collisionPair{id1: id2, id2: id1}
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
	cs.processedPairs = nil
}
