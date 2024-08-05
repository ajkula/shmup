package entities

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/ajkula/shmup/graphics"
	"github.com/hajimehoshi/ebiten/v2"
)

type Enemy struct {
	X, Y, Z float64
	Speed   float64
	Size    float64
	Color   color.Color
	RotX    float64
	RotY    float64
	RotZ    float64
	RotTri  float64
	Shape   int
	Active  bool
}

func NewEnemy(x, y float64) *Enemy {
	return &Enemy{
		X:     x,
		Y:     y,
		Z:     0,
		Speed: rand.Float64()*50 + 80, // Vitesse entre 50 et 100 pixels par seconde
		Size:  float64(rand.Intn(20) + 10),
		Color: color.RGBA{
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			uint8(rand.Intn(256)),
			255,
		},
		Shape: rand.Intn(3),
	}
}

func (e *Enemy) Update(deltaTime float64) {
	e.Y += e.Speed * deltaTime
	e.RotX += 0.02
	e.RotY += 0.03
	e.RotZ += 0.01
	e.RotTri += 0.05
	e.Z += math.Sin(e.Y/50) * deltaTime * 10
}

func (e *Enemy) Reset(screenWidth float64) {
	e.Y = -e.Size
	e.X = rand.Float64() * screenWidth
	e.RotX = 0
	e.RotY = 0
	e.RotZ = 0
}

func (e *Enemy) Draw(screen *ebiten.Image) {
	if e.Y+e.Size < 0 || e.Y > ScreenHeight+250 {
		return
	}
	scaleFactor := 1 + (e.Y / ScreenHeight)

	switch e.Shape {
	case 0:
		graphics.DrawTriangle(screen, e.X, e.Y, e.Size*scaleFactor, e.RotTri, e.Color)
	case 1:
		cubeSize := e.Size * scaleFactor
		graphics.DrawCube3D(screen, e.X, e.Y-cubeSize/2, e.Z, cubeSize, e.Color)
	case 2:
		graphics.DrawCircle(screen, e.X, e.Y, e.Size*scaleFactor, e.Color)
	}

	fmt.Printf("Drawing enemy at (%.2f, %.2f)\n", e.X, e.Y)
}

func NewEnemyFormation(pattern FormationPattern, count int, entryX, entryY, speed float64) *EnemyFormation {
	formation := &EnemyFormation{
		Pattern:           pattern,
		EntryPoint:        struct{ X, Y float64 }{entryX, entryY},
		Speed:             speed * 1.5,
		CurrentStep:       0,
		PauseTime:         0,
		FormationComplete: false,
		MaxEnemySize:      0,
	}

	for i := 0; i < count; i++ {
		var e *Enemy
		switch pattern {
		case RowPattern:
			x := float64(ScreenWidth)*0.1 + float64(i)*(float64(ScreenWidth)*0.8/float64(count-1))
			e = NewEnemy(x, -50-float64(i*10)) // Placer les ennemis juste au-dessus de l'écran
		case ColumnPattern:
			e = NewEnemy(entryX, -50-float64(i*30))
		case LoopPattern:
			e = NewEnemy(entryX, -100-float64(i*20))
		}
		e.Active = true
		formation.Enemies = append(formation.Enemies, e)

		// Mettre à jour la taille maximale des ennemis
		if e.Size > formation.MaxEnemySize {
			formation.MaxEnemySize = e.Size
		}
	}

	return formation
}
