package system

import (
	"context"
	"sync"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type RenderSystem struct {
	core.BaseSystem
	entities []types.Renderable
	mu       sync.Mutex
}

func NewRenderSystem() *RenderSystem {
	return &RenderSystem{
		entities: make([]types.Renderable, 0),
	}
}

func (rs *RenderSystem) Initialize(ctx context.Context) error {
	rs.CTX = ctx
	return nil
}

func (rs *RenderSystem) Update(deltaTime float64) error {
	// effectuer des opérations de préparation au rendu ici
	// rar exemple, trier les entités par ordre de profondeur, mettre à jour des animations...
	select {
	case <-rs.CTX.Done():
		return rs.CTX.Err()
	default:
		rs.mu.Lock()
		defer rs.mu.Unlock()
		// il faut déterminer quoi mettre en place
	}
	return nil
}

func (rs *RenderSystem) Run(ctx context.Context) error {
	return rs.BaseSystem.Run(ctx)
}

func (rs *RenderSystem) Shutdown() {
	// cleanup
}

func (rs *RenderSystem) Render(screen *ebiten.Image) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for _, entity := range rs.entities {
		entity.Draw(screen)
	}
}

func (rs *RenderSystem) AddEntity(entity types.Renderable) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.entities = append(rs.entities, entity)
}

func (rs *RenderSystem) RemoveEntity(entity types.Renderable) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for i, e := range rs.entities {
		if e == entity {
			rs.entities = append(rs.entities[:i], rs.entities[i+1:]...)
			break
		}
	}
}
