package game

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/event"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/manager"
	"github.com/ajkula/shmup/state"
	"github.com/ajkula/shmup/system"
	"github.com/ajkula/shmup/types"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	fixedDeltaTime = 1.0 / 60.0 // 60 fps
	maxDeltaTime   = 1.0 / 10.0 // max time between updates (10 fps)
)

type Game struct {
	ctx    context.Context
	cancel context.CancelFunc

	// sync
	inputSystem  *system.InputSystem
	renderSystem *system.RenderSystem

	// concurrents
	concurrentSystems []core.System

	player         *entity.Player
	lastUpdateTime time.Time
	accumulator    float64
	wg             sync.WaitGroup
	errChan        chan error
	eventManager   interfaces.EventManagerInterface
}

func NewGame(ctx context.Context) (*Game, error) {
	gameCtx, cancel := context.WithCancel(ctx)

	// EventManager - sera concurrent
	eventManager := event.NewEventManager()
	if err := eventManager.Initialize(gameCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize event manager: %w", err)
	}

	g := &Game{
		ctx:               gameCtx,
		cancel:            cancel,
		lastUpdateTime:    time.Now(),
		accumulator:       0,
		errChan:           make(chan error, 10),
		eventManager:      eventManager,
		concurrentSystems: make([]core.System, 0),
	}

	// Systèmes SYNCHRONES (appelés par Ebiten)
	g.inputSystem = system.NewInputSystem(eventManager)
	g.renderSystem = system.NewRenderSystem()

	// Initialize synchronous systems
	if err := g.inputSystem.Initialize(gameCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize input system: %w", err)
	}
	if err := g.renderSystem.Initialize(gameCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize render system: %w", err)
	}

	// Systèmes CONCURRENTS (tournent en arrière-plan)
	stateManager := state.NewStateManager(eventManager)
	collisionSystem := system.NewCollisionSystem(eventManager)
	enemyManager := manager.NewEnemyManager(eventManager)
	bulletManager := manager.NewBulletManager(eventManager)
	scoreManager := manager.NewScoreManager(eventManager)
	levelManager := manager.NewLevelManager(eventManager)

	g.concurrentSystems = append(g.concurrentSystems,
		eventManager, // EventManager tourne en concurrent
		stateManager,
		collisionSystem,
		enemyManager,
		bulletManager,
		scoreManager,
		levelManager,
	)

	// Initialize concurrent systems
	for _, sys := range g.concurrentSystems {
		if err := sys.Initialize(gameCtx); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize concurrent system: %w", err)
		}
	}

	// Create player
	g.player = entity.NewPlayer(
		types.Vector2D{
			X: float64(config.Config.ScreenWidth / 2),
			Y: float64(config.Config.ScreenHeight - 50),
		},
		eventManager,
	)

	// Add player to render system
	g.renderSystem.AddEntity(g.player)

	// Add some test enemies to see sprites
	enemy1 := entity.NewEnemyWithType(
		types.Vector2D{X: 100, Y: 50},
		eventManager,
		graphics.Scout,
		graphics.Level1,
	)
	enemy2 := entity.NewEnemyWithType(
		types.Vector2D{X: 200, Y: 80},
		eventManager,
		graphics.Scout,
		graphics.Level2,
	)
	enemy3 := entity.NewEnemyWithType(
		types.Vector2D{X: 300, Y: 50},
		eventManager,
		graphics.Fighter,
		graphics.Level1,
	)
	enemy4 := entity.NewEnemyWithType(
		types.Vector2D{X: 400, Y: 80},
		eventManager,
		graphics.Heavy,
		graphics.Level1,
	)

	g.renderSystem.AddEntity(enemy1)
	g.renderSystem.AddEntity(enemy2)
	g.renderSystem.AddEntity(enemy3)
	g.renderSystem.AddEntity(enemy4)

	g.startConcurrentSystems()

	return g, nil
}

func (g *Game) Update() error {
	select {
	case err := <-g.errChan:
		log.Printf("Concurrent system error: %v", err)
	default:
	}

	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	default:
	}

	// Calculate delta time
	currentTime := time.Now()
	deltaTime := currentTime.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = currentTime

	if deltaTime > maxDeltaTime {
		deltaTime = maxDeltaTime
	}

	g.accumulator += deltaTime

	// Fixed timestep updates
	for g.accumulator >= fixedDeltaTime {
		// Update synchronous systems
		if err := g.inputSystem.Update(fixedDeltaTime); err != nil {
			return fmt.Errorf("input system error: %w", err)
		}

		// Update player
		if err := g.player.Update(fixedDeltaTime); err != nil {
			return fmt.Errorf("player update error: %w", err)
		}

		g.accumulator -= fixedDeltaTime
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Clear screen to black
	screen.Fill(config.Config.BackgroundColor)

	g.renderSystem.Render(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.Config.ScreenWidth, config.Config.ScreenHeight
}

func (g *Game) startConcurrentSystems() {
	for _, sys := range g.concurrentSystems {
		g.wg.Add(1)
		go g.runConcurrentSystem(sys)
	}
}

func (g *Game) runConcurrentSystem(sys core.System) {
	defer g.wg.Done()

	if err := sys.Run(g.ctx); err != nil && err != context.Canceled {
		select {
		case g.errChan <- fmt.Errorf("concurrent system error: %w", err):
		default:
			log.Printf("Error channel full, logging: %v", err)
		}
	}
}

func (g *Game) Shutdown() {
	g.cancel()
	g.inputSystem.Shutdown()
	g.renderSystem.Shutdown()

	for _, sys := range g.concurrentSystems {
		sys.Shutdown()
	}

	g.wg.Wait()

	close(g.errChan)
}
