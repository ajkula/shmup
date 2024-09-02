package entity

import (
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Formation struct {
	types.BaseEntity
	enemies       []types.GameEntity
	formationType types.FormationType
	pattern       types.MovementPattern
	complete      bool
	eventManager  interfaces.EventManagerInterface
}

func NewFormation(formationType types.FormationType, pattern types.MovementPattern, position types.Vector2D, eventManager interfaces.EventManagerInterface) *Formation {
	return &Formation{
		BaseEntity: types.BaseEntity{
			Position: position,
		},
		enemies:       make([]types.GameEntity, 0),
		formationType: formationType,
		pattern:       pattern,
		complete:      false,
		eventManager:  eventManager,
	}
}

func (f *Formation) Update(deltaTime float64) error {
	if f.pattern != nil {
		formationMembers := make([]types.FormationMember, len(f.enemies))
		for i, enemy := range f.enemies {
			formationMembers[i] = enemy.(types.FormationMember)
		}
		f.pattern.Move(formationMembers, deltaTime)
	}
	f.checkCompletion()
	return nil
}

func (f *Formation) Draw(screen *ebiten.Image) {
	for _, enemy := range f.enemies {
		enemy.Draw(screen)
	}
}

func (f *Formation) GetEntities() []types.Entity {
	entities := make([]types.Entity, len(f.enemies))
	for i, enemy := range f.enemies {
		entities[i] = enemy
	}
	return entities
}

func (f *Formation) AddEntity(e types.Entity) {
	if enemy, ok := e.(types.GameEntity); ok {
		f.enemies = append(f.enemies, enemy)
		f.eventManager.Publish(interfaces.EnemyAddedToFormation, enemy)
	}
}

func (f *Formation) RemoveEntity(e types.Entity) {
	for i, enemy := range f.enemies {
		if enemy == e {
			f.enemies = append(f.enemies[:i], f.enemies[i+1:]...)
			f.eventManager.Publish(interfaces.EnemyRemovedFromFormation, enemy)
			break
		}
	}
}

func (f *Formation) IsComplete() bool {
	return f.complete
}

func (f *Formation) GetFormationType() types.FormationType {
	return f.formationType
}

func (f *Formation) SetPattern(pattern types.MovementPattern) {
	f.pattern = pattern
}

func (f *Formation) GetPattern() types.MovementPattern {
	return f.pattern
}

func (f *Formation) checkCompletion() {
	if !f.IsComplete() && len(f.enemies) == 0 {
		f.complete = true
		f.eventManager.Publish(interfaces.FormationDestroyed, f)
	}
}

var _ types.Formation = (*Formation)(nil)
