package entity

import (
	"math"
	"math/rand"

	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type SpiralPattern struct {
	BulletCount int
	BulletSpeed float64
	SpiralRate  float64
	TimeOffset  float64
}

func NewSpiralPattern(count int, speed float64, spiralRate float64) *SpiralPattern {
	return &SpiralPattern{
		BulletCount: count,
		BulletSpeed: speed,
		SpiralRate:  spiralRate,
		TimeOffset:  0,
	}
}

func (p *SpiralPattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	shots := make([]interfaces.BulletData, p.BulletCount)

	p.TimeOffset += 0.1
	angleStep := 2 * math.Pi / float64(p.BulletCount)

	for i := 0; i < p.BulletCount; i++ {
		angle := float64(i)*angleStep + p.TimeOffset*p.SpiralRate
		shots[i] = interfaces.BulletData{
			Position:  shooterPos,
			Direction: types.Vector2D{X: math.Cos(angle), Y: math.Sin(angle)},
			Speed:     p.BulletSpeed,
		}
	}
	return shots
}

type WavePattern struct {
	BulletCount int
	BulletSpeed float64
	WaveWidth   float64
	WaveHeight  float64
}

func NewWavePattern(count int, speed float64, width float64) *WavePattern {
	return &WavePattern{
		BulletCount: count,
		BulletSpeed: speed,
		WaveWidth:   width,
		WaveHeight:  0.3,
	}
}

func (p *WavePattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	shots := make([]interfaces.BulletData, p.BulletCount)

	for i := 0; i < p.BulletCount; i++ {
		offset := (float64(i)/float64(p.BulletCount-1) - 0.5) * p.WaveWidth
		angle := math.Atan2(1, offset*0.01)

		shots[i] = interfaces.BulletData{
			Position:  types.Vector2D{X: shooterPos.X + offset, Y: shooterPos.Y},
			Direction: types.Vector2D{X: math.Sin(angle) * p.WaveHeight, Y: math.Cos(angle)},
			Speed:     p.BulletSpeed,
		}
	}
	return shots
}

type SprayPattern struct {
	BulletCount int
	SpreadAngle float64
	BulletSpeed float64
}

func NewSprayPattern(count int, spreadDegrees float64, speed float64) *SprayPattern {
	return &SprayPattern{
		BulletCount: count,
		SpreadAngle: spreadDegrees * math.Pi / 180,
		BulletSpeed: speed,
	}
}

func (p *SprayPattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	shots := make([]interfaces.BulletData, p.BulletCount)

	baseAngle := math.Pi / 2 // Downward
	if targetPos != nil {
		dx := targetPos.X - shooterPos.X
		dy := targetPos.Y - shooterPos.Y
		baseAngle = math.Atan2(dy, dx)
	}

	angleStep := p.SpreadAngle / float64(p.BulletCount-1)
	startAngle := baseAngle - p.SpreadAngle/2

	for i := 0; i < p.BulletCount; i++ {
		angle := startAngle + float64(i)*angleStep
		shots[i] = interfaces.BulletData{
			Position:  shooterPos,
			Direction: types.Vector2D{X: math.Cos(angle), Y: math.Sin(angle)},
			Speed:     p.BulletSpeed,
		}
	}
	return shots
}

type BurstPattern struct {
	BulletCount int
	BulletSpeed float64
	Spread      float64
}

func NewBurstPattern(count int, speed float64, spreadDegrees float64) *BurstPattern {
	return &BurstPattern{
		BulletCount: count,
		BulletSpeed: speed,
		Spread:      spreadDegrees * math.Pi / 180,
	}
}

func (p *BurstPattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	shots := make([]interfaces.BulletData, p.BulletCount)

	baseAngle := math.Pi / 2
	if targetPos != nil {
		dx := targetPos.X - shooterPos.X
		dy := targetPos.Y - shooterPos.Y
		baseAngle = math.Atan2(dy, dx)
	}

	for i := 0; i < p.BulletCount; i++ {
		angle := baseAngle + (rand.Float64()-0.5)*p.Spread
		shots[i] = interfaces.BulletData{
			Position:  shooterPos,
			Direction: types.Vector2D{X: math.Cos(angle), Y: math.Sin(angle)},
			Speed:     p.BulletSpeed,
		}
	}
	return shots
}

type AimedPattern struct {
	BulletSpeed float64
}

func NewAimedPattern(speed float64) *AimedPattern {
	return &AimedPattern{BulletSpeed: speed}
}

func (p *AimedPattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	if targetPos == nil {
		return []interfaces.BulletData{{
			Position:  shooterPos,
			Direction: types.Vector2D{X: 0, Y: 1},
			Speed:     p.BulletSpeed,
		}}
	}

	dx := targetPos.X - shooterPos.X
	dy := targetPos.Y - shooterPos.Y
	length := math.Sqrt(dx*dx + dy*dy)

	if length == 0 {
		return []interfaces.BulletData{{
			Position:  shooterPos,
			Direction: types.Vector2D{X: 0, Y: 1},
			Speed:     p.BulletSpeed,
		}}
	}

	return []interfaces.BulletData{{
		Position:  shooterPos,
		Direction: types.Vector2D{X: dx / length, Y: dy / length},
		Speed:     p.BulletSpeed,
	}}
}

type CirclePattern struct {
	BulletCount int
	BulletSpeed float64
}

func NewCirclePattern(count int, speed float64) *CirclePattern {
	return &CirclePattern{
		BulletCount: count,
		BulletSpeed: speed,
	}
}

func (p *CirclePattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	shots := make([]interfaces.BulletData, p.BulletCount)
	angleStep := 2 * math.Pi / float64(p.BulletCount)

	for i := 0; i < p.BulletCount; i++ {
		angle := float64(i) * angleStep
		shots[i] = interfaces.BulletData{
			Position:  shooterPos,
			Direction: types.Vector2D{X: math.Cos(angle), Y: math.Sin(angle)},
			Speed:     p.BulletSpeed,
		}
	}
	return shots
}

type StreamPattern struct {
	BulletSpeed float64
}

func NewStreamPattern(speed float64) *StreamPattern {
	return &StreamPattern{BulletSpeed: speed}
}

func (p *StreamPattern) GenerateShots(shooterPos types.Vector2D, targetPos *types.Vector2D) []interfaces.BulletData {
	return []interfaces.BulletData{{
		Position:  shooterPos,
		Direction: types.Vector2D{X: 0, Y: 1},
		Speed:     p.BulletSpeed,
	}}
}
