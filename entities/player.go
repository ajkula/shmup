package entities

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Player struct {
	X, Y  float64
	Speed float64
	Size  float64
	Color color.Color
}

func NewPlayer(x, y float64) *Player {
	return &Player{
		X:     x,
		Y:     y,
		Speed: 200, // Vitesse en pixels par seconde
		Size:  20,
		Color: color.RGBA{0, 255, 0, 255}, // Green
	}
}

func (p *Player) Update(deltaTime, screenWidth, screenHeight float64) {
	moveSpeed := p.Speed * deltaTime

	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.X -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.X += moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.Y -= moveSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.Y += moveSpeed
	}

	p.X = math.Max(0, math.Min(p.X, screenWidth-p.Size))
	p.Y = math.Max(0, math.Min(p.Y, screenHeight-p.Size))
}

func (p *Player) Draw(screen *ebiten.Image) {
	vector.StrokeLine(screen, float32(p.X), float32(p.Y+p.Size), float32(p.X+p.Size), float32(p.Y+p.Size), 1, p.Color, true)
	vector.StrokeLine(screen, float32(p.X+p.Size), float32(p.Y+p.Size), float32(p.X+p.Size/2), float32(p.Y), 1, p.Color, true)
	vector.StrokeLine(screen, float32(p.X+p.Size/2), float32(p.Y), float32(p.X), float32(p.Y+p.Size), 1, p.Color, true)
}
