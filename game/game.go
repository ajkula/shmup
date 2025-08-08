package game

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/event"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/manager"
	"github.com/ajkula/shmup/registry"
	"github.com/ajkula/shmup/state"
	"github.com/ajkula/shmup/system"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	fixedDeltaTime = 1.0 / 60.0
	maxDeltaTime   = 1.0 / 10.0
)

type Game struct {
	ctx    context.Context
	cancel context.CancelFunc

	synchronousSystems []core.System
	eventDrivenSystems []core.System

	player         *entity.Player
	lastUpdateTime time.Time
	accumulator    float64
	errChan        chan error
	eventManager   interfaces.EventManagerInterface

	inputSystem    *system.InputSystem
	renderSystem   *system.RenderSystem
	entityRegistry *registry.EntityRegistry
}

func NewGame(ctx context.Context) (*Game, error) {
	gameCtx, cancel := context.WithCancel(ctx)

	eventManager := event.NewEventManager()
	if err := eventManager.Initialize(gameCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize event manager: %w", err)
	}

	g := &Game{
		ctx:                gameCtx,
		cancel:             cancel,
		lastUpdateTime:     time.Now(),
		accumulator:        0,
		errChan:            make(chan error, 10),
		eventManager:       eventManager,
		synchronousSystems: make([]core.System, 0),
		eventDrivenSystems: make([]core.System, 0),
	}

	g.inputSystem = system.NewInputSystem(eventManager)
	g.renderSystem = system.NewRenderSystem()
	g.entityRegistry = registry.NewEntityRegistry(eventManager)
	systemTicker := system.NewSystemTicker(eventManager)

	g.renderSystem.SetEntityRegistry(g.entityRegistry)

	g.synchronousSystems = append(g.synchronousSystems,
		g.inputSystem,
		g.renderSystem,
		g.entityRegistry,
		systemTicker,
	)

	stateManager := state.NewStateManager(eventManager)
	collisionSystem := system.NewCollisionSystem(eventManager)
	enemyManager := manager.NewEnemyManager(eventManager)
	bulletManager := manager.NewBulletManager(eventManager)
	scoreManager := manager.NewScoreManager(eventManager)
	levelManager := manager.NewLevelManager(eventManager)
	waveManager := manager.NewWaveManager(eventManager)

	g.entityRegistry.RegisterProvider(enemyManager)
	g.entityRegistry.RegisterProvider(bulletManager)

	g.eventDrivenSystems = append(g.eventDrivenSystems,
		stateManager,
		collisionSystem,
		enemyManager,
		bulletManager,
		scoreManager,
		levelManager,
		waveManager,
	)

	for _, sys := range g.synchronousSystems {
		if err := sys.Initialize(gameCtx); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize synchronous system: %w", err)
		}
	}

	for _, sys := range g.eventDrivenSystems {
		if err := sys.Initialize(gameCtx); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize event-driven system: %w", err)
		}
	}

	g.player = entity.NewPlayer(
		types.Vector2D{
			X: float64(config.Config.ScreenWidth / 2),
			Y: float64(config.Config.ScreenHeight - 50),
		},
		eventManager,
	)

	g.entityRegistry.AddStaticEntity(g.player)

	return g, nil
}

func (g *Game) Update() error {
	select {
	case err := <-g.errChan:
		log.Printf("System error: %v", err)
	default:
	}

	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	default:
	}

	currentTime := time.Now()
	deltaTime := currentTime.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = currentTime

	if deltaTime > maxDeltaTime {
		deltaTime = maxDeltaTime
	}

	g.accumulator += deltaTime

	for g.accumulator >= fixedDeltaTime {
		// Update synchronous systems
		for _, sys := range g.synchronousSystems {
			if err := sys.Update(fixedDeltaTime); err != nil {
				return fmt.Errorf("synchronous system error: %w", err)
			}
		}

		// Update event-driven systems
		for _, sys := range g.eventDrivenSystems {
			if err := sys.Update(fixedDeltaTime); err != nil {
				return fmt.Errorf("event-driven system error: %w", err)
			}
		}

		if err := g.player.Update(fixedDeltaTime); err != nil {
			return fmt.Errorf("player update error: %w", err)
		}

		g.accumulator -= fixedDeltaTime
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(config.Config.BackgroundColor)
	g.renderSystem.Render(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.Config.ScreenWidth, config.Config.ScreenHeight
}

func (g *Game) Shutdown() {
	g.cancel()

	g.inputSystem.Shutdown()
	g.renderSystem.Shutdown()
	g.entityRegistry.Shutdown()

	for _, sys := range g.synchronousSystems[3:] { // Skip input, render, and entity registry
		sys.Shutdown()
	}

	for _, sys := range g.eventDrivenSystems {
		sys.Shutdown()
	}

	g.eventManager.Shutdown()

	close(g.errChan)
}