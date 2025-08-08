package system

import (
	"context"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type EntityRegistry interface {
	GetSnapshot() []types.Renderable
}

type RenderSystem struct {
	core.BaseSystem
	entityRegistry EntityRegistry
}

func NewRenderSystem() *RenderSystem {
	return &RenderSystem{}
}

func (rs *RenderSystem) SetEntityRegistry(registry EntityRegistry) {
	rs.entityRegistry = registry
}

func (rs *RenderSystem) Initialize(ctx context.Context) error {
	return rs.BaseSystem.Initialize(ctx)
}

func (rs *RenderSystem) Update(deltaTime float64) error {
	select {
	case <-rs.CTX.Done():
		return rs.CTX.Err()
	default:
		return nil
	}
}

func (rs *RenderSystem) Shutdown() {
	rs.BaseSystem.Shutdown()
}

func (rs *RenderSystem) Render(screen *ebiten.Image) {
	if rs.entityRegistry == nil {
		return
	}

	entities := rs.entityRegistry.GetSnapshot()
	for _, entity := range entities {
		entity.Draw(screen)
	}
}
