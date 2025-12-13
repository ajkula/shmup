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

// DiamondFormationPattern - Diamond/Rhombus shape (classic arcade)
type DiamondFormationPattern struct {
	BaseMovementPattern
	Spacing float64
}

func NewDiamondFormationPattern(config PatternConfig) *DiamondFormationPattern {
	return &DiamondFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Spacing: config.Spacing,
	}
}

func (d *DiamondFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Diamond pattern: top point, sides, bottom point
	// Works best with 5, 9, 13 enemies
	layers := int(math.Sqrt(float64(enemyCount)))
	centerLayer := layers / 2

	layer := 0
	indexInLayer := enemyIndex

	for layer < layers {
		enemiesInLayer := layer*2 + 1
		if indexInLayer < enemiesInLayer {
			break
		}
		indexInLayer -= enemiesInLayer
		layer++
	}

	layerCenter := float64(layer*2) / 2.0
	xOffset := (float64(indexInLayer) - layerCenter) * d.Spacing
	yOffset := float64(layer-centerLayer) * d.Spacing * 0.8

	return types.Vector2D{X: xOffset, Y: yOffset}
}

func (d *DiamondFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{X: 0, Y: d.VerticalSpeed * time}
}

// WingsFormationPattern - Two groups flanking from sides (Galaga-style)
type WingsFormationPattern struct {
	BaseMovementPattern
	Spacing    float64
	WingSpread float64
}

func NewWingsFormationPattern(config PatternConfig) *WingsFormationPattern {
	return &WingsFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Spacing:    config.Spacing,
		WingSpread: config.Radius, // Reuse radius for wing spread
	}
}

func (w *WingsFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Split into two wings
	halfCount := enemyCount / 2
	isLeftWing := enemyIndex < halfCount

	wingIndex := enemyIndex
	if !isLeftWing {
		wingIndex -= halfCount
	}

	xBase := -w.WingSpread
	if !isLeftWing {
		xBase = w.WingSpread
	}

	return types.Vector2D{
		X: xBase + (float64(wingIndex%3) * w.Spacing * 0.5),
		Y: float64(wingIndex/3) * w.Spacing * -0.6,
	}
}

func (w *WingsFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{X: 0, Y: w.VerticalSpeed * time}
}

// SpiralFormationPattern - Enemies in spiral (modern shmup)
type SpiralFormationPattern struct {
	BaseMovementPattern
	Radius        float64
	SpiralTight   float64
	RotationSpeed float64
}

func NewSpiralFormationPattern(config PatternConfig) *SpiralFormationPattern {
	return &SpiralFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Radius:        config.Radius,
		SpiralTight:   config.Spacing,
		RotationSpeed: config.RotationSpeed,
	}
}

func (sp *SpiralFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Spiral outward from center
	angle := float64(enemyIndex) * (2 * math.Pi / 5) // 5 enemies per rotation
	radius := sp.Radius + (float64(enemyIndex) * sp.SpiralTight)
	currentAngle := angle + (sp.RotationSpeed * time)

	return types.Vector2D{
		X: math.Cos(currentAngle) * radius,
		Y: math.Sin(currentAngle) * radius,
	}
}

func (sp *SpiralFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{X: 0, Y: sp.VerticalSpeed * time}
}

// ArrowFormationPattern - Arrow/Wedge pointing down (aggressive)
type ArrowFormationPattern struct {
	BaseMovementPattern
	Spacing float64
}

func NewArrowFormationPattern(config PatternConfig) *ArrowFormationPattern {
	return &ArrowFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Spacing: config.Spacing,
	}
}

func (a *ArrowFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	// Arrow pointing down: leader at front, rows behind
	row := int(math.Sqrt(float64(enemyIndex)))
	indexInRow := enemyIndex - (row * row)
	rowWidth := row*2 + 1

	centerPos := float64(rowWidth-1) / 2.0
	xOffset := (float64(indexInRow) - centerPos) * a.Spacing
	yOffset := float64(row) * a.Spacing * -0.7 // Negative = point forward

	return types.Vector2D{X: xOffset, Y: yOffset}
}

func (a *ArrowFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{X: 0, Y: a.VerticalSpeed * time}
}

// ZigZagFormationPattern - Zigzag pattern (classic arcade)
type ZigZagFormationPattern struct {
	BaseMovementPattern
	Amplitude float64
	Spacing   float64
	Phase     float64
}

func NewZigZagFormationPattern(config PatternConfig) *ZigZagFormationPattern {
	return &ZigZagFormationPattern{
		BaseMovementPattern: BaseMovementPattern{
			Speed:         config.Speed,
			VerticalSpeed: config.VerticalSpeed,
		},
		Amplitude: config.Amplitude,
		Spacing:   config.Spacing,
		Phase:     config.Frequency,
	}
}

func (z *ZigZagFormationPattern) GetOffset(enemyIndex int, time float64, formation types.FormationController) types.Vector2D {
	enemyCount := len(formation.GetEnemies())
	if enemyCount == 0 {
		return types.Vector2D{}
	}

	centerIndex := float64(enemyCount-1) / 2.0
	relativeIndex := float64(enemyIndex) - centerIndex

	// Zigzag based on index and time
	zigzag := math.Sin((relativeIndex+time*z.Phase)*0.5) * z.Amplitude

	return types.Vector2D{
		X: relativeIndex*z.Spacing + zigzag,
		Y: 0,
	}
}

func (z *ZigZagFormationPattern) GetCenterMovement(time float64) types.Vector2D {
	return types.Vector2D{X: 0, Y: z.VerticalSpeed * time}
}
