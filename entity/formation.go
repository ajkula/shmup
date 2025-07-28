package entity

import (
	"math"

	"github.com/ajkula/shmup/types"
)

type FormationPattern func(t float64) types.Vector2D

type Formation struct {
	ID        string
	EnemyIDs  []string
	Pattern   FormationPattern
	CenterPos types.Vector2D
	Time      float64
}

func NewFormation(id string, pattern FormationPattern, centerPos types.Vector2D) *Formation {
	return &Formation{
		ID:        id,
		EnemyIDs:  make([]string, 0),
		Pattern:   pattern,
		CenterPos: centerPos,
		Time:      0,
	}
}

func (f *Formation) AddEnemy(enemyID string) {
	f.EnemyIDs = append(f.EnemyIDs, enemyID)
}

func (f *Formation) RemoveEnemy(enemyID string) {
	for i, id := range f.EnemyIDs {
		if id == enemyID {
			f.EnemyIDs = append(f.EnemyIDs[:i], f.EnemyIDs[i+1:]...)
			break
		}
	}
}

// Getters and setters
func (f *Formation) GetID() string                   { return f.ID }
func (f *Formation) GetEnemyIDs() []string           { return f.EnemyIDs }
func (f *Formation) GetPattern() FormationPattern    { return f.Pattern }
func (f *Formation) GetCenterPos() types.Vector2D    { return f.CenterPos }
func (f *Formation) SetCenterPos(pos types.Vector2D) { f.CenterPos = pos }
func (f *Formation) GetTime() float64                { return f.Time }
func (f *Formation) SetTime(t float64)               { f.Time = t }

var (
	CirclePattern = func(radius float64) FormationPattern {
		return func(t float64) types.Vector2D {
			return types.Vector2D{
				X: math.Cos(t) * radius,
				Y: math.Sin(t) * radius,
			}
		}
	}

	SineWavePattern = func(amplitude, frequency float64) FormationPattern {
		return func(t float64) types.Vector2D {
			return types.Vector2D{
				X: t * frequency,
				Y: math.Sin(t) * amplitude,
			}
		}
	}

	VFormationPattern = func(spacing float64) FormationPattern {
		return func(t float64) types.Vector2D {
			return types.Vector2D{
				X: spacing * math.Cos(math.Pi/6),
				Y: spacing * math.Sin(math.Pi/6),
			}
		}
	}
)

// Fonction pour créer des formations spécifiques
func NewCircleFormation(id string, centerPos types.Vector2D, radius float64) *Formation {
	return NewFormation(id, CirclePattern(radius), centerPos)
}

func NewSineWaveFormation(id string, centerPos types.Vector2D, amplitude, frequency float64) *Formation {
	return NewFormation(id, SineWavePattern(amplitude, frequency), centerPos)
}

func NewVFormation(id string, centerPos types.Vector2D, spacing float64) *Formation {
	return NewFormation(id, VFormationPattern(spacing), centerPos)
}
