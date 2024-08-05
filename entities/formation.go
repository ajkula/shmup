package entities

import (
	"fmt"
	"math"
	"time"
)

const (
	ScreenWidth  = 520
	ScreenHeight = 800
)

type FormationPattern int

const (
	ColumnPattern FormationPattern = iota
	RowPattern    FormationPattern = iota
	LoopPattern   FormationPattern = iota
)

type EnemyFormation struct {
	Enemies           []*Enemy
	Pattern           FormationPattern
	EntryPoint        struct{ X, Y float64 }
	Speed             float64
	Time              float64
	PauseTime         time.Duration
	LoopPoint         float64
	CurrentStep       int
	Width             float64
	Height            float64
	SpawnInterval     float64
	LastSpawnTime     float64
	SpawnedEnemies    int
	FormationComplete bool
	MaxEnemySize      float64
}

func (f *EnemyFormation) Update(deltaTime float64) bool {
	switch f.Pattern {
	case ColumnPattern:
		f.updateColumnPattern(deltaTime)
	case RowPattern:
		f.updateRowPattern(deltaTime)
	case LoopPattern:
		f.updateLoopPattern(deltaTime)
	}

	allOffscreen := true
	for _, e := range f.Enemies {
		if e.Y < ScreenHeight+e.Size+200 { // Ajout d'une marge de 50 pixels
			allOffscreen = false
			break
		}
	}

	f.FormationComplete = allOffscreen
	fmt.Printf("Formation update: Pattern %v, Enemies: %d, All offscreen: %v\n", f.Pattern, len(f.Enemies), allOffscreen)
	return allOffscreen
}

func (f *EnemyFormation) updateColumnPattern(deltaTime float64) {
	spacing := float64(ScreenHeight) * 0.8 / float64(len(f.Enemies)-1)
	startY := float64(ScreenHeight) * 0.1

	for i, e := range f.Enemies {
		targetY := startY + float64(i)*spacing
		if e.Y < targetY {
			e.Y += f.Speed * deltaTime * 2
			if e.Y > targetY {
				e.Y = targetY
			}
		} else {
			e.Y += f.Speed * deltaTime
		}
		e.X = f.EntryPoint.X
	}

	fmt.Printf("Column pattern update: First enemy Y: %.2f, Last enemy Y: %.2f\n",
		f.Enemies[0].Y, f.Enemies[len(f.Enemies)-1].Y)
}

func (f *EnemyFormation) updateRowPattern(deltaTime float64) {
	totalWidth := float64(ScreenWidth) * 0.8
	spacing := totalWidth / float64(len(f.Enemies)-1)
	startX := float64(ScreenWidth) * 0.1
	targetY := 50.0 // Position Y cible initiale pour la rangée

	allInPosition := true

	for i, e := range f.Enemies {
		targetX := startX + float64(i)*spacing

		// Mouvement vers la position horizontale cible
		if e.X < targetX {
			e.X += f.Speed * deltaTime
			if e.X > targetX {
				e.X = targetX
			}
			allInPosition = false
		} else if e.X > targetX {
			e.X -= f.Speed * deltaTime
			if e.X < targetX {
				e.X = targetX
			}
			allInPosition = false
		}

		// Mouvement vertical
		if e.Y < targetY {
			e.Y += f.Speed * deltaTime
			allInPosition = false
		} else if allInPosition {
			e.Y += f.Speed * deltaTime / 2
		}
	}

	fmt.Printf("Row pattern update: First enemy position: (%.2f, %.2f), allInPosition: %v\n",
		f.Enemies[0].X, f.Enemies[0].Y, allInPosition)
}

func (f *EnemyFormation) updateLoopPattern(deltaTime float64) {
	f.Time += deltaTime
	radius := float64(ScreenWidth) * 0.2
	centerX := float64(ScreenWidth) / 2
	centerY := -100 + f.Time*f.Speed // Fait descendre le centre progressivement

	for i, e := range f.Enemies {
		angle := f.Time*2 + float64(i)*(2*math.Pi/float64(len(f.Enemies)))
		e.X = centerX + radius*math.Cos(angle)
		e.Y = centerY + radius*math.Sin(angle)

		// Garder les ennemis dans les limites horizontales de l'écran
		e.X = math.Max(0, math.Min(e.X, ScreenWidth))
	}

	fmt.Printf("Loop pattern update: Center Y: %.2f, First enemy position: (%.2f, %.2f)\n", centerY, f.Enemies[0].X, f.Enemies[0].Y)
}

func (f *EnemyFormation) checkDestroyed() bool {
	for _, e := range f.Enemies {
		if e.Active {
			return false
		}
	}
	return true
}
