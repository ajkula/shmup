package game

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/effects"
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

	player           *entity.Player
	lastUpdateTime   time.Time
	accumulator      float64
	errChan          chan error
	eventManager     interfaces.EventManagerInterface
	explosionManager *effects.ExplosionManager

	inputSystem      *system.InputSystem
	renderSystem     *system.RenderSystem
	uiSystem         *system.UISystem
	entityRegistry   *registry.EntityRegistry
	stateManager     *state.StateManager
	scoreManager     *manager.ScoreManager
	waveManager      *manager.WaveManager
	enemyManager     *manager.EnemyManager
	bulletManager    *manager.BulletManager
	anyKeyWasPressed  bool
	gameOverStartTime time.Time
	gameOverDelay     float64
	victoryStartTime  time.Time
	victoryDelay      float64
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
		gameOverDelay:      3.0, // 3 seconds delay before accepting input
		victoryDelay:       3.0, // 3 seconds delay before accepting input
	}

	g.explosionManager = effects.NewExplosionManager()
	explosionSystem := system.NewExplosionSystem(g.explosionManager, eventManager)

	g.inputSystem = system.NewInputSystem(eventManager)
	g.renderSystem = system.NewRenderSystem()
	g.uiSystem = system.NewUISystem()
	g.entityRegistry = registry.NewEntityRegistry(eventManager)
	systemTicker := system.NewSystemTicker(eventManager)

	g.renderSystem.SetEntityRegistry(g.entityRegistry)

	g.synchronousSystems = append(g.synchronousSystems,
		g.inputSystem,
		g.renderSystem,
		g.entityRegistry,
		systemTicker,
	)

	g.stateManager = state.NewStateManager(eventManager)
	collisionSystem := system.NewCollisionSystem(eventManager)
	g.enemyManager = manager.NewEnemyManager(eventManager)
	g.bulletManager = manager.NewBulletManager(eventManager)
	g.scoreManager = manager.NewScoreManager(eventManager)
	levelManager := manager.NewLevelManager(eventManager)
	g.waveManager = manager.NewWaveManager(eventManager)

	g.player = entity.NewPlayer(
		types.Vector2D{
			X: float64(config.Config.ScreenWidth / 2),
			Y: float64(config.Config.ScreenHeight - 50),
		},
		eventManager,
	)

	g.entityRegistry.AddStaticEntity(g.player)
	g.bulletManager.SetPlayer(g.player)
	collisionSystem.SetPlayer(g.player)

	g.entityRegistry.RegisterProvider(g.enemyManager)
	g.entityRegistry.RegisterProvider(g.bulletManager)
	collisionSystem.SetEntityRegistry(g.entityRegistry)

	g.eventDrivenSystems = append(g.eventDrivenSystems,
		g.stateManager,
		collisionSystem,
		g.enemyManager,
		g.bulletManager,
		g.scoreManager,
		levelManager,
		g.waveManager,
		explosionSystem,
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

	g.entityRegistry.AddStaticEntity(g.player)

	// Setup game-level event listeners
	g.setupEventListeners()

	return g, nil
}

func (g *Game) setupEventListeners() {
	// Player died - switch to game over if no lives left
	playerDiedCh, err := g.eventManager.Subscribe(interfaces.PlayerDestroyed)
	if err == nil {
		go func() {
			for range playerDiedCh {
				// Don't trigger Game Over if we're already in Victory state
				currentState := g.stateManager.GetState()
				if currentState == state.StateVictory {
					fmt.Println("[Game] Player destroyed but Victory already triggered - ignoring")
					continue
				}
				fmt.Println("[Game] Player destroyed - Game Over!")
				g.gameOverStartTime = time.Now()
				g.anyKeyWasPressed = true // Reset to prevent immediate key detection
				g.stateManager.SetState(state.StateGameOver)
			}
		}()
	}

	// Victory - all levels completed
	victoryCh, err := g.eventManager.Subscribe(interfaces.Victory)
	if err == nil {
		go func() {
			for range victoryCh {
				fmt.Println("[Game] Victory! All levels completed!")
				g.victoryStartTime = time.Now()
				g.anyKeyWasPressed = true // Reset to prevent immediate key detection
				g.stateManager.SetState(state.StateVictory)
			}
		}()
	}
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

	// Handle state transitions based on input
	currentState := g.stateManager.GetState()

	// Update UI system (for animations)
	currentTime := time.Now()
	deltaTime := currentTime.Sub(g.lastUpdateTime).Seconds()
	g.lastUpdateTime = currentTime

	if deltaTime > maxDeltaTime {
		deltaTime = maxDeltaTime
	}

	g.uiSystem.Update(deltaTime, currentState)

	// Check for any key press
	anyKeyNow := g.isAnyKeyPressed()
	justPressed := anyKeyNow && !g.anyKeyWasPressed
	g.anyKeyWasPressed = anyKeyNow

	switch currentState {
	case state.StateMainMenu:
		if justPressed {
			g.startGame()
		}
		return nil
	case state.StateGameOver:
		// Only accept input after 3 seconds delay
		timeSinceGameOver := time.Since(g.gameOverStartTime).Seconds()
		if justPressed && timeSinceGameOver >= g.gameOverDelay {
			g.resetGame()
		}
		return nil
	case state.StateVictory:
		// Only accept input after 3 seconds delay
		timeSinceVictory := time.Since(g.victoryStartTime).Seconds()
		if justPressed && timeSinceVictory >= g.victoryDelay {
			g.resetGame()
		}
		return nil
	}

	// Only update game logic when playing
	if currentState != state.StatePlaying {
		return nil
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

		g.explosionManager.Update(fixedDeltaTime)
		g.accumulator -= fixedDeltaTime
	}

	return nil
}

func (g *Game) isAnyKeyPressed() bool {
	// Check common keys
	keys := []ebiten.Key{
		ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyEscape,
		ebiten.KeyA, ebiten.KeyB, ebiten.KeyC, ebiten.KeyD, ebiten.KeyE,
		ebiten.KeyF, ebiten.KeyG, ebiten.KeyH, ebiten.KeyI, ebiten.KeyJ,
		ebiten.KeyK, ebiten.KeyL, ebiten.KeyM, ebiten.KeyN, ebiten.KeyO,
		ebiten.KeyP, ebiten.KeyQ, ebiten.KeyR, ebiten.KeyS, ebiten.KeyT,
		ebiten.KeyU, ebiten.KeyV, ebiten.KeyW, ebiten.KeyX, ebiten.KeyY,
		ebiten.KeyZ,
		ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyArrowLeft, ebiten.KeyArrowRight,
	}

	for _, key := range keys {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}
	return false
}

func (g *Game) startGame() {
	fmt.Println("[Game] Starting game")
	g.stateManager.SetState(state.StatePlaying)
}

func (g *Game) resetGame() {
	fmt.Println("[Game] Resetting game")

	// Clear all game entities directly (no events to avoid deadlock)
	g.bulletManager.ClearAllBullets()
	g.enemyManager.ClearAllEnemies()

	// Reset player
	g.player.Reset()

	// Reset score
	g.scoreManager.ResetScore()

	// Reset wave manager (this will clear all waves and reset to level 1)
	g.waveManager.ResetToLevel1()

	// Clear all explosions
	g.explosionManager = effects.NewExplosionManager()

	// Return to main menu
	g.stateManager.SetState(state.StateMainMenu)

	fmt.Println("[Game] Game reset complete - back to menu")
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(config.Config.BackgroundColor)

	// Only render game entities when playing
	if g.stateManager.GetState() == state.StatePlaying {
		g.renderSystem.Render(screen)
		g.explosionManager.Draw(screen)
	}

	// Always draw UI (handles all states: menu, playing, gameover, victory)
	// Get player health
	playerHealth := g.player.GetHealth()
	playerMaxHealth := g.player.GetMaxHealth()

	// Get boss health if active
	boss := g.enemyManager.GetBoss()
	bossActive := boss != nil
	bossHealth := 0
	bossMaxHealth := 1000 // Default if no boss
	if bossActive {
		bossHealth = boss.GetHealth()
		bossMaxHealth = boss.GetMaxHealth()
	}

	g.uiSystem.Draw(screen, g.stateManager.GetState(), g.player.GetLives(), g.scoreManager.GetScore(), g.waveManager.GetCurrentLevel(), g.scoreManager.GetHighScore(), playerHealth, playerMaxHealth, bossActive, bossHealth, bossMaxHealth)
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
