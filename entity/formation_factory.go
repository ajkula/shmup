package entity

import (
	"fmt"

	"github.com/ajkula/shmup/config"
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
	cfg := PatternConfig{
		Speed:         60.0,
		VerticalSpeed: config.Config.EnemySpeed,
	}

	switch formationType {
	case PresetVFormation:
		cfg.Spacing = 40.0
		pattern = NewVFormationPattern(cfg)
	case PresetLineFormation:
		cfg.Spacing = 50.0
		cfg.VerticalSpeed = config.Config.EnemySpeed * 1.2 // a bit faster
		pattern = NewLineFormationPattern(cfg)
	case PresetCircleFormation:
		cfg.Radius = 60.0
		cfg.RotationSpeed = 1.0
		cfg.VerticalSpeed = config.Config.EnemySpeed * 0.8 // a bit slower
		pattern = NewCircleFormationPattern(cfg)
	case PresetSineWaveFormation:
		cfg.Amplitude = 100.0
		cfg.Frequency = 2.0
		cfg.Spacing = 45.0
		cfg.VerticalSpeed = config.Config.EnemySpeed * 0.9
		pattern = NewSineWaveFormationPattern(cfg)
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

func (ff *FormationFactory) CreateFormationWithConfig(
	formationType PresetFormationType,
	centerPos types.Vector2D,
	enemyType graphics.EnemyType,
	enemyLevel graphics.EnemyLevel,
	enemyCount int,
	patternConfig PatternConfig,
) types.FormationController {
	ff.idCounter++
	id := fmt.Sprintf("formation_%d", ff.idCounter)

	var pattern types.MovementPattern
	switch formationType {
	case PresetVFormation:
		pattern = NewVFormationPattern(patternConfig)
	case PresetLineFormation:
		pattern = NewLineFormationPattern(patternConfig)
	case PresetCircleFormation:
		pattern = NewCircleFormationPattern(patternConfig)
	case PresetSineWaveFormation:
		pattern = NewSineWaveFormationPattern(patternConfig)
	}

	formation := NewFormation(id, pattern, centerPos, ff.eventManager)

	for i := 0; i < enemyCount; i++ {
		offset := pattern.GetOffset(i, 0, formation)
		enemyPos := centerPos.Add(offset)

		enemy := NewEnemyWithType(enemyPos, ff.eventManager, enemyType, enemyLevel)
		formation.AddEnemy(enemy)

		ff.eventManager.Publish(interfaces.EnemyCreated, enemy)
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

	if len(enemies) > 0 {
		for _, enemy := range enemies {
			formation.AddEnemy(enemy)
			ff.eventManager.Publish(interfaces.EnemyAddedToFormation, map[string]any{
				"enemy":     enemy,
				"formation": formation,
			})
		}
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

func (ff *FormationFactory) SpawnFormation(
	pattern types.MovementPattern,
	centerPos types.Vector2D,
	enemyCount int,
) types.FormationController {
	enemies := make([]types.GameEntity, 0, enemyCount)
	formation := NewFormation("temp", pattern, centerPos, ff.eventManager)

	for i := 0; i < enemyCount; i++ {
		offset := pattern.GetOffset(i, 0, formation)
		enemyPos := centerPos.Add(offset)

		enemy := NewEnemyWithType(enemyPos, ff.eventManager, graphics.Scout, graphics.Level1)
		enemies = append(enemies, enemy)
		ff.eventManager.Publish(interfaces.EnemyCreated, enemy)
	}

	return ff.CreateCustomFormation(pattern, centerPos, enemies)
}

func (ff *FormationFactory) SpawnFormationWithType(
	pattern types.MovementPattern,
	centerPos types.Vector2D,
	enemyType graphics.EnemyType,
	enemyLevel graphics.EnemyLevel,
	enemyCount int,
) types.FormationController {
	enemies := make([]types.GameEntity, 0, enemyCount)
	formation := NewFormation("temp", pattern, centerPos, ff.eventManager)

	for i := 0; i < enemyCount; i++ {
		offset := pattern.GetOffset(i, 0, formation)
		enemyPos := centerPos.Add(offset)

		enemy := NewEnemyWithType(enemyPos, ff.eventManager, enemyType, enemyLevel)
		enemies = append(enemies, enemy)
		ff.eventManager.Publish(interfaces.EnemyCreated, enemy)
	}

	return ff.CreateCustomFormation(pattern, centerPos, enemies)
}
