package entity

import (
	"math"

	"github.com/ajkula/shmup/types"
)

// SOLID: Open/Closed Principle
// New patterns can be added without modifying existing code

// BaseMovementPattern provides common functionality
type BaseMovementPattern struct {
	Speed         float64
	VerticalSpeed float64
}

type PatternConfig struct {
	Spacing       float64
	Speed         float64
	VerticalSpeed float64
	Radius        float64 // CirclePattern
	Amplitude     float64 // SineWavePattern
	Frequency     float64 // SineWavePattern
	RotationSpeed float64 // CirclePattern
}

// VFormationPattern - Classic V formation moving down
type VFormationPattern struct {
	BaseMovementPattern
	Spacing     float64
	AngleOffset float64
}

func NewVFormationPattern(config PatternConfig) *VFormationPattern {
	return &VFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Spacing:     config.Spacing,
		AngleOffset: math.Pi / 6,
	}
}

func (v *VFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Center the formation around index 0
	centerIndex := float64(enemyCount-1) / 2.0
	relativeIndex := float64(enemyIndex) - centerIndex

	xOffset := relativeIndex * v.Spacing
	yOffset := math.Abs(relativeIndex) * v.Spacing * 0.5

	return types.Vector2D{
		X: xOffset,
		Y: -yOffset,
	}
}

func (v *VFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{
		X: 0,
		Y: v.VerticalSpeed * time,
	}
}

// LineFormationPattern - Horizontal line moving down
type LineFormationPattern struct {
	BaseMovementPattern
	Spacing float64
}

func NewLineFormationPattern(config PatternConfig) *LineFormationPattern {
	return &LineFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Spacing: config.Spacing,
	}
}

func (l *LineFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	centerIndex := float64(enemyCount-1) / 2.0
	relativeIndex := float64(enemyIndex) - centerIndex

	return types.Vector2D{
		X: relativeIndex * l.Spacing,
		Y: 0,
	}
}

func (l *LineFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{
		X: 0,
		Y: l.VerticalSpeed * time,
	}
}

// CircleFormationPattern - Enemies in rotating circle
type CircleFormationPattern struct {
	BaseMovementPattern
	Radius        float64
	RotationSpeed float64
}

func NewCircleFormationPattern(config PatternConfig) *CircleFormationPattern {
	return &CircleFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Radius:        config.Radius,
		RotationSpeed: config.RotationSpeed,
	}
}

func (c *CircleFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Distribute enemies evenly around circle
	angleStep := 2 * math.Pi / float64(enemyCount)
	baseAngle := float64(enemyIndex) * angleStep
	currentAngle := baseAngle + (c.RotationSpeed * time)

	return types.Vector2D{
		X: math.Cos(currentAngle) * c.Radius,
		Y: math.Sin(currentAngle) * c.Radius,
	}
}

func (c *CircleFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{
		X: 0,
		Y: c.VerticalSpeed * time,
	}
}

// SineWaveFormationPattern - Formation follows sine wave
type SineWaveFormationPattern struct {
	BaseMovementPattern
	Amplitude float64
	Frequency float64
	Spacing   float64
}

func NewSineWaveFormationPattern(config PatternConfig) *SineWaveFormationPattern {
	return &SineWaveFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Amplitude: config.Amplitude,
		Frequency: config.Frequency,
		Spacing:   config.Spacing,
	}
}

func (s *SineWaveFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	centerIndex := float64(enemyCount-1) / 2.0
	relativeIndex := float64(enemyIndex) - centerIndex

	return types.Vector2D{
		X: relativeIndex * s.Spacing,
		Y: 0,
	}
}

func (s *SineWaveFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	sineOffset := math.Sin(time*s.Frequency) * s.Amplitude

	return types.Vector2D{
		X: sineOffset,
		Y: s.VerticalSpeed * time,
	}
}
