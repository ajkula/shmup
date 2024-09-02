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
	ctx            context.Context
	cancel         context.CancelFunc
	systems        []core.System
	lastUpdateTime time.Time
	accumulator    float64
	player         *entity.Player
	wg             sync.WaitGroup
	updateChan     chan float64
	errChan        chan error
}

func NewGame(ctx context.Context) (*Game, error) {
	gameCtx, cancel := context.WithCancel(ctx)

	eventManager := event.NewEventManager()
	if err := eventManager.Initialize(gameCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize event manager: %w", err)
	}
	stateManager := state.NewStateManager(eventManager)

	g := &Game{
		ctx:            gameCtx,
		cancel:         cancel,
		systems:        make([]core.System, 0),
		lastUpdateTime: time.Now(),
		accumulator:    0,
		updateChan:     make(chan float64, 1),
		errChan:        make(chan error, 1),
	}

	// initialize systems
	renderSystem := system.NewRenderSystem()
	collisionSystem := system.NewCollisionSystem(eventManager)
	inputSystem := system.NewInputSystem(eventManager)
	updateSystem := system.NewUpdateSystem()

	// initialize managers
	enemyManager := manager.NewEnemyManager(eventManager)
	bulletManager := manager.NewBulletManager(eventManager)
	scoreManager := manager.NewScoreManager(eventManager)
	levelManager := manager.NewLevelManager(eventManager)

	// add all systems and managers to the game
	g.systems = append(g.systems,
		eventManager,
		stateManager,
		renderSystem,
		collisionSystem,
		inputSystem,
		updateSystem,
		enemyManager,
		bulletManager,
		scoreManager,
		levelManager,
	)

	// initialize all systems
	for _, sys := range g.systems {
		if err := sys.Initialize(gameCtx); err != nil {
			cancel()
			return nil, fmt.Errorf("failed to initialize system: %w", err)
		}
	}

	// create player
	g.player = entity.NewPlayer(
		types.Vector2D{
			X: float64(config.Config.ScreenWidth / 2),
			Y: float64(config.Config.ScreenHeight - 50),
		},
		eventManager,
	)

	return g, nil
}

func (g *Game) Update() error {
	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	case err := <-g.errChan:
		g.Shutdown()
		return fmt.Errorf("critical error occurred: %w", err)
	default:
		currentTime := time.Now()
		deltaTime := currentTime.Sub(g.lastUpdateTime).Seconds()
		g.lastUpdateTime = currentTime

		if deltaTime > maxDeltaTime {
			deltaTime = maxDeltaTime
		}

		g.accumulator += deltaTime

		for g.accumulator >= fixedDeltaTime {
			select {
			case <-g.ctx.Done():
				return g.ctx.Err()
			default:
				g.updateChan <- fixedDeltaTime
				g.accumulator -= fixedDeltaTime
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if renderSystem, ok := g.systems[2].(*system.RenderSystem); ok {
		renderSystem.Render(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.Config.ScreenWidth, config.Config.ScreenHeight
}

func (g *Game) Run() {
	g.wg.Add(len(g.systems))
	for _, sys := range g.systems {
		go func(s core.System) {
			defer g.wg.Done()

			for {
				select {
				case <-g.ctx.Done():
					return
				case dt := <-g.updateChan:
					if err := s.Update(dt); err != nil {
						if err != context.Canceled {
							log.Printf("Error running system: %v", err)
						}
						g.handleCriticalError(err)
						return
					}
				}
			}

		}(sys)
	}
}

func (g *Game) handleCriticalError(err error) {
	select {
	case g.errChan <- err:
	default:
		log.Printf("Critical error occurred but error channel is full: %v", err)
	}
	g.cancel()
}

func (g *Game) Shutdown() {
	g.cancel()
	for _, sys := range g.systems {
		sys.Shutdown()
	}
	g.wg.Wait()
	close(g.updateChan)
	close(g.errChan)
}
