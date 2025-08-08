package entity

import (
	"fmt"

	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

// SOLID: Dependency Inversion + Factory Pattern
// High-level modules don't depend on low-level modules

type FormationFactory struct {
	eventManager interfaces.EventManagerInterface
	idCounter    int
}

func NewFormationFactory(eventManager interfaces.EventManagerInterface) *FormationFactory {
	return &FormationFactory{
		eventManager: eventManager,
		idCounter:    0,
	}
}

// PresetFormationType for easy creation
type PresetFormationType int

const (
	PresetVFormation PresetFormationType = iota
	PresetLineFormation
	PresetCircleFormation
	PresetSineWaveFormation
)

func (ff *FormationFactory) CreatePresetFormation(
	formationType PresetFormationType,
	centerPos types.Vector2D,
	enemyType graphics.EnemyType,
	enemyLevel graphics.EnemyLevel,
	enemyCount int,
) types.FormationController {
	ff.idCounter++
	id := fmt.Sprintf("formation_%d", ff.idCounter)

	var pattern types.MovementPattern
	switch formationType {
	case PresetVFormation:
		pattern = NewVFormationPattern(40.0, 60.0)
	case PresetLineFormation:
		pattern = NewLineFormationPattern(50.0, 80.0)
	case PresetCircleFormation:
		pattern = NewCircleFormationPattern(60.0, 1.0, 70.0)
	case PresetSineWaveFormation:
		pattern = NewSineWaveFormationPattern(100.0, 2.0, 45.0, 65.0)
	default:
		pattern = NewVFormationPattern(40.0, 60.0)
	}

	formation := NewFormation(id, pattern, centerPos, ff.eventManager)

	for i := 0; i < enemyCount; i++ {
		offset := pattern.GetOffset(i, 0, formation) // Offset
		enemyPos := centerPos.Add(offset)            // Position

		enemy := NewEnemyWithType(enemyPos, ff.eventManager, enemyType, enemyLevel)
		formation.AddEnemy(enemy)

		ff.eventManager.Publish(interfaces.EnemyCreated, enemy)
		ff.eventManager.Publish(interfaces.EnemyAddedToFormation, map[string]interface{}{
			"enemy":     enemy,
			"formation": formation,
		})
	}

	return formation
}

// CreateCustomFormation for full control
func (ff *FormationFactory) CreateCustomFormation(
	pattern types.MovementPattern,
	centerPos types.Vector2D,
	enemies []types.GameEntity,
) types.FormationController {

	ff.idCounter++
	id := fmt.Sprintf("formation_%d", ff.idCounter)

	formation := NewFormation(id, pattern, centerPos, ff.eventManager)

	for _, enemy := range enemies {
		formation.AddEnemy(enemy)
		ff.eventManager.Publish(interfaces.EnemyAddedToFormation, map[string]interface{}{
			"enemy":     enemy,
			"formation": formation,
		})
	}

	return formation
}

// Quick creation methods for common scenarios
func (ff *FormationFactory) CreateScoutVFormation(centerPos types.Vector2D, count int) types.FormationController {
	return ff.CreatePresetFormation(PresetVFormation, centerPos, graphics.Scout, graphics.Level1, count)
}

func (ff *FormationFactory) CreateFighterLine(centerPos types.Vector2D, count int) types.FormationController {
	return ff.CreatePresetFormation(PresetLineFormation, centerPos, graphics.Fighter, graphics.Level1, count)
}

func (ff *FormationFactory) CreateHeavyCircle(centerPos types.Vector2D, count int) types.FormationController {
	return ff.CreatePresetFormation(PresetCircleFormation, centerPos, graphics.Heavy, graphics.Level1, count)
}

func (ff *FormationFactory) CreateMixedWave(centerPos types.Vector2D) []types.FormationController {
	formations := make([]types.FormationController, 0, 3)

	// Lead scouts in V formation
	scoutPos := types.Vector2D{X: centerPos.X, Y: centerPos.Y - 100}
	formations = append(formations, ff.CreateScoutVFormation(scoutPos, 5))

	// Fighter line behind
	fighterPos := types.Vector2D{X: centerPos.X, Y: centerPos.Y}
	formations = append(formations, ff.CreateFighterLine(fighterPos, 4))

	// Heavy escort circle
	heavyPos := types.Vector2D{X: centerPos.X + 150, Y: centerPos.Y + 50}
	formations = append(formations, ff.CreateHeavyCircle(heavyPos, 3))

	return formations
}

// SpawnFormation implements FormationSpawner interface
func (ff *FormationFactory) SpawnFormation(pattern types.MovementPattern, centerPos types.Vector2D, enemyCount int) types.FormationController {
	return ff.CreateCustomFormation(pattern, centerPos, make([]types.GameEntity, 0, enemyCount))
}
