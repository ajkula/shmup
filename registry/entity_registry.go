package registry

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type EntityProvider interface {
	GetRenderableEntities() []types.Renderable
}

type EntityRegistry struct {
	core.BaseSystem
	providers      []EntityProvider
	staticEntities []types.Renderable
	snapshot       atomic.Value
	mu             sync.RWMutex
	eventManager   interfaces.EventManagerInterface
	isShutdown     int32
}

func NewEntityRegistry(eventManager interfaces.EventManagerInterface) *EntityRegistry {
	registry := &EntityRegistry{
		providers:      make([]EntityProvider, 0),
		staticEntities: make([]types.Renderable, 0),
		eventManager:   eventManager,
	}

	registry.snapshot.Store([]types.Renderable{})
	return registry
}

func (er *EntityRegistry) Initialize(ctx context.Context) error {
	err := er.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	ch, err := er.eventManager.Subscribe(interfaces.SystemTick)
	if err != nil {
		return err
	}

	go er.eventListener(ch)
	return nil
}

func (er *EntityRegistry) Update(deltaTime float64) error {
	select {
	case <-er.CTX.Done():
		return er.CTX.Err()
	default:
		return nil
	}
}

func (er *EntityRegistry) eventListener(systemTickCh <-chan interfaces.Event) {
	for {
		select {
		case <-er.CTX.Done():
			return
		case _, ok := <-systemTickCh:
			if !ok {
				return
			}
			if atomic.LoadInt32(&er.isShutdown) == 0 {
				er.updateSnapshot()
			}
		}
	}
}

func (er *EntityRegistry) GetRenderableEntities() []types.Renderable {
	return er.GetSnapshot()
}

func (er *EntityRegistry) updateSnapshot() {
	er.mu.RLock()

	totalEntities := make([]types.Renderable, 0, len(er.staticEntities)+len(er.providers)*10)

	for _, entity := range er.staticEntities {
		totalEntities = append(totalEntities, entity)
	}

	for _, provider := range er.providers {
		entities := provider.GetRenderableEntities()
		totalEntities = append(totalEntities, entities...)
	}

	er.mu.RUnlock()

	er.snapshot.Store(totalEntities)
}

func (er *EntityRegistry) RegisterProvider(provider EntityProvider) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.providers = append(er.providers, provider)
}

func (er *EntityRegistry) AddStaticEntity(entity types.Renderable) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.staticEntities = append(er.staticEntities, entity)
}

func (er *EntityRegistry) RemoveStaticEntity(entity types.Renderable) {
	er.mu.Lock()
	defer er.mu.Unlock()
	for i, e := range er.staticEntities {
		if e == entity {
			er.staticEntities = append(er.staticEntities[:i], er.staticEntities[i+1:]...)
			break
		}
	}
}

func (er *EntityRegistry) GetSnapshot() []types.Renderable {
	snapshot := er.snapshot.Load()
	if entities, ok := snapshot.([]types.Renderable); ok {
		return entities
	}
	return []types.Renderable{}
}

func (er *EntityRegistry) Shutdown() {
	if !atomic.CompareAndSwapInt32(&er.isShutdown, 0, 1) {
		return
	}

	er.mu.Lock()
	defer er.mu.Unlock()

	er.providers = nil
	er.staticEntities = nil
	er.snapshot.Store([]types.Renderable{})
}

var _ core.System = (*EntityRegistry)(nil)
