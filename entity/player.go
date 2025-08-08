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
	ShootCooldown float64
	eventManager  interfaces.EventManagerInterface
	inputEvents   <-chan interfaces.Event
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
			Position: position,
			Width:    32, Height: 32,
			Speed:  5,
			Health: 100,
			Color:  color.RGBA{0, 255, 0, 255},
		},
		ShootCooldown: 0,
		eventManager:  eventManager,
		inputEvents:   inputChan,
	}

	if config.Config.PlayerSpeed > 0 {
		player.Speed = config.Config.PlayerSpeed
	}

	return player
}

func (p *Player) Update(deltaTime float64) error {
	p.ShootCooldown -= deltaTime
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
	p.TakeDamage(10)
	p.eventManager.Publish(interfaces.PlayerDamaged, p)
	if p.Health <= 0 {
		p.eventManager.Publish(interfaces.PlayerDestroyed, p)
	}
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
