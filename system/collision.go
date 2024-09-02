package system

import (
	"context"
	"sync"
	"time"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

const (
	fixedDeltaTime = 1.0 / 60.0
)

type CollisionSystem struct {
	core.BaseSystem
	collidables   []types.GameEntity
	mu            sync.Mutex
	accumulator   float64
	lastCheckTime time.Time
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
}

func NewCollisionSystem(eventManager interfaces.EventManagerInterface) *CollisionSystem {
	return &CollisionSystem{
		collidables:   make([]types.GameEntity, 0),
		lastCheckTime: time.Now(),
		eventManager:  eventManager,
	}
}

func (cs *CollisionSystem) Initialize(ctx context.Context) error {
	err := cs.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	eventTypes := []interfaces.EventType{
		interfaces.BulletCreated,
		interfaces.BulletDestroyed,
		interfaces.EnemyDestroyed,
		interfaces.BossDefeated,
		interfaces.PlayerDestroyed,
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
		cs.mu.Lock()
		defer cs.mu.Unlock()

		currentTime := time.Now()
		actualDeltaTime := currentTime.Sub(cs.lastCheckTime).Seconds()
		cs.lastCheckTime = currentTime
		cs.accumulator += actualDeltaTime

		cs.processEvents()

		for cs.accumulator >= fixedDeltaTime {
			cs.CheckCollisions(fixedDeltaTime)
			cs.accumulator -= fixedDeltaTime
		}
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
			if entity, ok := evt.Data.(types.GameEntity); ok {
				switch eventType {
				case interfaces.BulletCreated:
					cs.AddCollidable(entity)
				default:
					cs.RemoveCollidable(entity)
				}
			}
		default:
			// nothing
		}
	}
}

func (cs *CollisionSystem) Run(ctx context.Context) error {
	return cs.BaseSystem.Run(ctx)
}

func (cs *CollisionSystem) Shutdown() {
	cs.collidables = nil
}

func (cs *CollisionSystem) AddCollidable(c types.GameEntity) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.collidables = append(cs.collidables, c)
}

func (cs *CollisionSystem) RemoveCollidable(c types.GameEntity) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for i, collidable := range cs.collidables {
		if collidable == c {
			cs.collidables = append(cs.collidables[:i], cs.collidables[i+1:]...)
			break
		}
	}
}

func (cs *CollisionSystem) CheckCollisions(deltaTime float64) {
	for i := 0; i < len(cs.collidables); i++ {
		for j := i + 1; j < len(cs.collidables); j++ {
			if cs.collidables[i].CanCollideWith(cs.collidables[j]) {
				if cs.detectCollision(cs.collidables[i], cs.collidables[j]) {
					cs.collidables[i].OnCollision(cs.collidables[j])
					cs.collidables[j].OnCollision(cs.collidables[i])
				}
			}
		}
	}
}

func (cs *CollisionSystem) detectCollision(a, b types.Entity) bool {
	ax, ay, aw, ah := a.GetCollisionBox()
	bx, by, bw, bh := b.GetCollisionBox()

	return ax < bx+bw &&
		ax+aw > bx &&
		ay < by+bh &&
		ay+ah > by
}
