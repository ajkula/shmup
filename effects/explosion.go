package effects

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type PixelParticle struct {
	Position      types.Vector2D
	Velocity      types.Vector2D
	Color         color.RGBA
	Life          float64 // 0.0 à 1.0
	MaxDistance   float64
	StartPos      types.Vector2D
	TotalDistance float64
}

type PixelExplosion struct {
	particles []PixelParticle
	alive     bool
	duration  float64
	elapsed   float64
}

type ExplosionManager struct {
	explosions []*PixelExplosion
}

func NewExplosionManager() *ExplosionManager {
	return &ExplosionManager{
		explosions: make([]*PixelExplosion, 0),
	}
}

func (em *ExplosionManager) CreateExplosionFromSprite(sprite *graphics.SpriteData, position types.Vector2D) {
	if sprite == nil || sprite.Pixels == nil {
		return
	}

	explosion := &PixelExplosion{
		particles: make([]PixelParticle, 0),
		alive:     true,
		duration:  1.0,
		elapsed:   0,
	}

	spriteSize := math.Sqrt(float64(sprite.Width*sprite.Width + sprite.Height*sprite.Height))
	maxDistance := spriteSize * 2.0 * 4.0

	for y := 0; y < sprite.Height; y++ {
		for x := 0; x < sprite.Width; x++ {
			pixelColor := sprite.Pixels[y][x]

			if rgba, ok := pixelColor.(color.RGBA); ok && rgba.A == 0 {
				continue
			}

			pixelWorldPos := types.Vector2D{
				X: position.X + float64(x*4),
				Y: position.Y + float64(y*4),
			}

			centerX := position.X + float64(sprite.Width*2)
			centerY := position.Y + float64(sprite.Height*2)

			dx := pixelWorldPos.X - centerX
			dy := pixelWorldPos.Y - centerY

			length := math.Sqrt(dx*dx + dy*dy)
			if length < 0.01 {
				length = 1
			}

			dirX := (dx/length + (rand.Float64()-0.5)*0.5)
			dirY := (dy/length + (rand.Float64()-0.5)*0.5)

			newLength := math.Sqrt(dirX*dirX + dirY*dirY)
			if newLength > 0 {
				dirX /= newLength
				dirY /= newLength
			}

			speed := 50.0 + rand.Float64()*150.0 // 50-200 pixels/sec

			particle := PixelParticle{
				Position: pixelWorldPos,
				StartPos: pixelWorldPos,
				Velocity: types.Vector2D{
					X: dirX * speed,
					Y: dirY * speed,
				},
				Color:         pixelColor.(color.RGBA),
				Life:          1.0,
				MaxDistance:   maxDistance,
				TotalDistance: 0,
			}

			explosion.particles = append(explosion.particles, particle)
		}
	}

	em.explosions = append(em.explosions, explosion)
}

func (em *ExplosionManager) CreateExplosionFromEntity(entity types.Entity) {
	position := entity.GetPosition()
	sprite := entity.GetSprite()

	if sprite != nil {
		em.CreateExplosionFromSprite(sprite, position)
	} else {
		em.CreateGenericExplosion(position, entity.GetColor())
	}
}

func (em *ExplosionManager) CreateGenericExplosion(position types.Vector2D, baseColor color.Color) {
	explosion := &PixelExplosion{
		particles: make([]PixelParticle, 0),
		alive:     true,
		duration:  1.0,
		elapsed:   0,
	}

	particleCount := 20 + rand.Intn(10)
	maxDistance := 100.0

	rgba := color.RGBAModel.Convert(baseColor).(color.RGBA)

	for i := 0; i < particleCount; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := 50.0 + rand.Float64()*150.0

		particle := PixelParticle{
			Position: position,
			StartPos: position,
			Velocity: types.Vector2D{
				X: math.Cos(angle) * speed,
				Y: math.Sin(angle) * speed,
			},
			Color:         rgba,
			Life:          1.0,
			MaxDistance:   maxDistance,
			TotalDistance: 0,
		}

		explosion.particles = append(explosion.particles, particle)
	}

	em.explosions = append(em.explosions, explosion)
}

func (em *ExplosionManager) Update(deltaTime float64) {
	aliveExplosions := make([]*PixelExplosion, 0)

	for _, explosion := range em.explosions {
		explosion.Update(deltaTime)
		if explosion.alive {
			aliveExplosions = append(aliveExplosions, explosion)
		}
	}

	em.explosions = aliveExplosions
}

func (em *ExplosionManager) Draw(screen *ebiten.Image) {
	for _, explosion := range em.explosions {
		explosion.Draw(screen)
	}
}

func (e *PixelExplosion) Update(deltaTime float64) {
	e.elapsed += deltaTime

	if e.elapsed >= e.duration {
		e.alive = false
		return
	}

	for i := range e.particles {
		p := &e.particles[i]

		oldPos := p.Position
		p.Position.X += p.Velocity.X * deltaTime
		p.Position.Y += p.Velocity.Y * deltaTime

		dx := p.Position.X - oldPos.X
		dy := p.Position.Y - oldPos.Y
		p.TotalDistance += math.Sqrt(dx*dx + dy*dy)

		progress := p.TotalDistance / p.MaxDistance
		if progress > 0.5 {
			fadeProgress := (progress - 0.5) * 2.0
			p.Life = 1.0 - fadeProgress
			if p.Life < 0 {
				p.Life = 0
			}
		}

		p.Velocity.X *= 0.98
		p.Velocity.Y *= 0.98
	}
}

func (e *PixelExplosion) Draw(screen *ebiten.Image) {
	for _, p := range e.particles {
		if p.Life <= 0 {
			continue
		}

		alpha := uint8(float64(p.Color.A) * p.Life)
		fadedColor := color.RGBA{
			R: p.Color.R,
			G: p.Color.G,
			B: p.Color.B,
			A: alpha,
		}

		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				screen.Set(
					int(p.Position.X)+dx,
					int(p.Position.Y)+dy,
					fadedColor,
				)
			}
		}
	}
}
