package system

import (
	"context"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/types"
)

type UpdateSystem struct {
	core.BaseSystem
	entities []types.Updatable
	mu       sync.Mutex
}

func NewUpdateSystem() *UpdateSystem {
	return &UpdateSystem{
		entities: make([]types.Updatable, 0),
	}
}

func (us *UpdateSystem) Initialize(ctx context.Context) error {
	us.CTX = ctx
	return nil
}

func (us *UpdateSystem) Update(deltaTime float64) error {
	select {
	case <-us.CTX.Done():
		return us.CTX.Err()
	default:
		us.mu.Lock()
		defer us.mu.Unlock()
		for _, entity := range us.entities {
			if err := entity.Update(deltaTime); err != nil {
				return err
			}
		}
		return nil
	}
}

func (us *UpdateSystem) Run(ctx context.Context) error {
	return us.BaseSystem.Run(ctx)
}

func (us *UpdateSystem) Shutdown() {
	// cleanup
}

func (us *UpdateSystem) AddEntity(entity types.Updatable) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.entities = append(us.entities, entity)
}

func (us *UpdateSystem) RemoveEntity(entity types.Updatable) {
	us.mu.Lock()
	defer us.mu.Unlock()
	for i, e := range us.entities {
		if e == entity {
			us.entities = append(us.entities[:i], us.entities[i+1:]...)
			break
		}
	}
}
