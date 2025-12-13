package entity

import (
	"image/color"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	types.BaseEntity
	playerClass              graphics.PlayerClass
	ShootCooldown            float64
	eventManager             interfaces.EventManagerInterface
	inputEvents              <-chan interfaces.Event
	Lives                    int
	MaxLives                 int
	InvulnerabilityTimer     float64
	RespawnInvulnerability   float64
}

func NewPlayer(position types.Vector2D, eventManager interfaces.EventManagerInterface) *Player {
	inputChan, err := eventManager.Subscribe(interfaces.InputEvent)
	if err != nil {
		dummyChan := make(chan interfaces.Event, 1)
		close(dummyChan)
		inputChan = dummyChan
	}

	player := &Player{
		BaseEntity: types.BaseEntity{
			Position:  position,
			Width:     32, Height: 32,
			Speed:     5,
			Health:    100,
			MaxHealth: 100,
			Color:     color.RGBA{0, 255, 0, 255},
		},
		ShootCooldown:          0,
		eventManager:           eventManager,
		inputEvents:            inputChan,
		playerClass:            graphics.StandardFighter,
		Lives:                  3,
		MaxLives:               3,
		InvulnerabilityTimer:   0,
		RespawnInvulnerability: 2.0,
	}

	if config.Config.PlayerSpeed > 0 {
		player.Speed = config.Config.PlayerSpeed
	}

	return player
}

func (p *Player) Update(deltaTime float64) error {
	p.ShootCooldown -= deltaTime

	// Update invulnerability timer
	if p.InvulnerabilityTimer > 0 {
		p.InvulnerabilityTimer -= deltaTime
	}

	p.processInputEvents()
	p.constrainToScreen()
	return nil
}

func (p *Player) processInputEvents() {
	for {
		select {
		case event := <-p.inputEvents:
			if data, ok := event.Data.(map[string]interface{}); ok {
				switch data["type"] {
				case "movement":
					p.handleMovement(data)
				case "shoot":
					p.Shoot()
				}
			}
		default:
			return
		}
	}
}

func (p *Player) GetSprite() *graphics.SpriteData {
	return graphics.GetPlayerSprite(graphics.StandardFighter)
}

func (p *Player) handleMovement(data map[string]interface{}) {
	deltaTime, ok := data["delta_time"].(float64)
	if !ok {
		return
	}

	moveSpeed := p.Speed * deltaTime * 60

	if left, ok := data["left"].(bool); ok && left {
		p.Position.X -= moveSpeed
	}
	if right, ok := data["right"].(bool); ok && right {
		p.Position.X += moveSpeed
	}
	if up, ok := data["up"].(bool); ok && up {
		p.Position.Y -= moveSpeed
	}
	if down, ok := data["down"].(bool); ok && down {
		p.Position.Y += moveSpeed
	}
}

func (p *Player) constrainToScreen() {
	if p.Position.X < 0 {
		p.Position.X = 0
	}
	if p.Position.X > float64(config.Config.ScreenWidth)-p.Width {
		p.Position.X = float64(config.Config.ScreenWidth) - p.Width
	}
	if p.Position.Y < 0 {
		p.Position.Y = 0
	}
	if p.Position.Y > float64(config.Config.ScreenHeight)-p.Height {
		p.Position.Y = float64(config.Config.ScreenHeight) - p.Height
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	if !p.IsAlive() {
		return // Don't draw dead player
	}

	// Blink during invulnerability
	if p.InvulnerabilityTimer > 0 {
		// Blink every 0.1 seconds
		blinkCycle := int(p.InvulnerabilityTimer * 10)
		if blinkCycle%2 == 0 {
			return // Skip drawing to create blink effect
		}
	}

	sprite := graphics.GetPlayerSprite(graphics.StandardFighter)
	graphics.DrawSprite(screen, sprite, p.Position.X, p.Position.Y)
}

func (p *Player) CanCollideWith(other types.Entity) bool {
	switch o := other.(type) {
	case *Enemy:
		return true
	case *Boss:
		return true
	case *Bullet:
		return o.IsEnemyBullet()
	default:
		return false
	}
}

func (p *Player) OnCollision(other types.Entity) {
	// Skip damage if invulnerable
	if p.InvulnerabilityTimer > 0 {
		return
	}

	switch o := other.(type) {
	case *Enemy:
		p.TakeDamage(20)
		p.eventManager.Publish(interfaces.PlayerDamaged, p)
	case *Boss:
		p.TakeDamage(50)
		p.eventManager.Publish(interfaces.PlayerDamaged, p)
	case *Bullet:
		if o.IsEnemyBullet() {
			p.TakeDamage(10)
			p.eventManager.Publish(interfaces.PlayerDamaged, p)
		}
	}

	if p.Health <= 0 {
		p.Die()
	}
}

func (p *Player) Die() {
	p.Lives--

	if p.eventManager != nil {
		p.eventManager.Publish(interfaces.PlayerDied, map[string]any{
			"lives": p.Lives,
		})
	}

	if p.Lives <= 0 {
		// Game Over
		if p.eventManager != nil {
			p.eventManager.Publish(interfaces.PlayerDestroyed, p)
		}
	} else {
		// Respawn
		p.Respawn()
	}
}

func (p *Player) Respawn() {
	p.Health = 100
	p.Position.X = float64(config.Config.ScreenWidth/2) - p.Width/2
	p.Position.Y = float64(config.Config.ScreenHeight - 100)
	p.InvulnerabilityTimer = p.RespawnInvulnerability
}

func (p *Player) GetLives() int {
	return p.Lives
}

func (p *Player) Reset() {
	p.Lives = p.MaxLives
	p.Health = p.MaxHealth
	p.InvulnerabilityTimer = 0
	p.ShootCooldown = 0
	p.Position.X = float64(config.Config.ScreenWidth/2) - p.Width/2
	p.Position.Y = float64(config.Config.ScreenHeight - 100)
}

func (p *Player) GetPlayerClass() graphics.PlayerClass {
	return p.playerClass
}

func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0
}

func (p *Player) Shoot() {
	if p.CanShoot() {
		p.eventManager.Publish(interfaces.PlayerShot, p)
		p.ShootCooldown = 0.2
	}
}

var _ types.GameEntity = (*Player)(nil)
