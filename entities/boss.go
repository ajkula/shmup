package entities

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Boss struct {
	X, Y           float64
	Speed          float64
	Size           float64
	Color          color.Color
	HP             int
	Patterns       []func(*Boss, float64)
	CurrentPattern int
	Time           float64
}

func NewBoss(x, y float64) *Boss {
	b := &Boss{
		X:              x,
		Y:              y,
		Speed:          2,
		Size:           50,
		Color:          color.RGBA{255, 0, 0, 255}, // Rouge
		HP:             100,
		Time:           0,
		CurrentPattern: 0,
	}
	b.Patterns = []func(*Boss, float64){
		bossPatternZigZag,
		bossPatternCircle,
		// Ajoutez d'autres patterns ici
	}
	return b
}

func (b *Boss) Update(deltaTime float64) {
	b.Time += deltaTime
	if len(b.Patterns) > 0 {
		b.Patterns[b.CurrentPattern](b, deltaTime)
	}
	// Changer de pattern toutes les 10 secondes, par exemple
	if int(b.Time/10)%len(b.Patterns) != b.CurrentPattern {
		b.CurrentPattern = int(b.Time/10) % len(b.Patterns)
	}
}

func (b *Boss) Draw(screen *ebiten.Image) {
	vector.DrawFilledCircle(screen, float32(b.X), float32(b.Y), float32(b.Size/2), b.Color, true)
}

func (b *Boss) TakeDamage(damage int) {
	b.HP -= damage
	if b.HP < 0 {
		b.HP = 0
	}
}

func (b *Boss) IsDead() bool {
	return b.HP <= 0
}

func bossPatternZigZag(b *Boss, deltaTime float64) {
	b.X += math.Sin(b.Time) * b.Speed
	b.Y += 0.5 * b.Speed
	b.X = math.Max(b.Size/2, math.Min(b.X, float64(ScreenWidth)-b.Size/2))
	b.Y = math.Max(b.Size/2, math.Min(b.Y, float64(ScreenHeight)-b.Size/2))
}

func bossPatternCircle(b *Boss, deltaTime float64) {
	centerX, centerY := float64(ScreenWidth)/2, float64(ScreenHeight)/3
	radius := 100.0
	b.X = centerX + math.Cos(b.Time*b.Speed/50)*radius
	b.Y = centerY + math.Sin(b.Time*b.Speed/50)*radius
}

// Ajoutez d'autres patterns de mouvement ici
