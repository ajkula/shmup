package entity

import (
	"math"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/types"
)

// SOLID: Open/Closed Principle
// New patterns can be added without modifying existing code

// BaseMovementPattern provides common functionality
type BaseMovementPattern struct {
	Speed         float64
	VerticalSpeed float64
}

// VFormationPattern - Classic V formation moving down
type VFormationPattern struct {
	BaseMovementPattern
	Spacing     float64
	AngleOffset float64
}

func NewVFormationPattern(spacing, speed float64) *VFormationPattern {
	return &VFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         speed,
			VerticalSpeed: config.Config.EnemySpeed,
		},
		Spacing:     spacing,
		AngleOffset: math.Pi / 6, // 30 degrees
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

	// V formation: enemies spread in V shape
	side := 1.0
	if relativeIndex < 0 {
		side = -1.0
	}

	distance := math.Abs(relativeIndex) * v.Spacing

	return types.Vector2D{
		X: side * distance * math.Cos(v.AngleOffset),
		Y: -distance * math.Sin(v.AngleOffset), // Slight upward offset for V
	}
}

func (v *VFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{
		X: 0,
		Y: v.VerticalSpeed * time * 60, // Add 60x multiplier
	}
}

// LineFormationPattern - Horizontal line moving down
type LineFormationPattern struct {
	BaseMovementPattern
	Spacing float64
}

func NewLineFormationPattern(spacing, speed float64) *LineFormationPattern {
	return &LineFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         speed,
			VerticalSpeed: config.Config.EnemySpeed,
		},
		Spacing: spacing,
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

func NewCircleFormationPattern(radius, rotationSpeed, speed float64) *CircleFormationPattern {
	return &CircleFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         speed,
			VerticalSpeed: config.Config.EnemySpeed,
		},
		Radius:        radius,
		RotationSpeed: rotationSpeed,
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

func NewSineWaveFormationPattern(amplitude, frequency, spacing, speed float64) *SineWaveFormationPattern {
	return &SineWaveFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         speed,
			VerticalSpeed: config.Config.EnemySpeed,
		},
		Amplitude: amplitude,
		Frequency: frequency,
		Spacing:   spacing,
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
