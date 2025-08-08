package manager

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

// WaveDefinition defines a wave to be spawned
type WaveDefinition struct {
	ID            string
	FormationType entity.PresetFormationType
	EnemyType     graphics.EnemyType
	EnemyLevel    graphics.EnemyLevel
	EnemyCount    int
	SpawnDelay    float64 // seconds after previous wave
	SpawnPosition types.Vector2D
	Priority      int // higher = spawns first if multiple ready
}

// WaveState tracks wave lifecycle
type WaveState int

const (
	WavePending   WaveState = iota // Waiting to spawn
	WaveActive                     // Formation is active on screen
	WaveCompleted                  // All enemies destroyed or exited
)

// ActiveWave tracks a spawned wave
type ActiveWave struct {
	Definition *WaveDefinition
	Formation  types.FormationController
	State      WaveState
	SpawnTime  float64
}

// WaveManager manages dynamic multi-wave spawning for arcade shmup
type WaveManager struct {
	core.BaseSystem

	// Threading
	mu            sync.RWMutex
	eventManager  interfaces.EventManagerInterface
	eventChannels map[interfaces.EventType]<-chan interfaces.Event
	isShutdown    int32

	// Wave management
	pendingWaves   []WaveDefinition
	activeWaves    []*ActiveWave
	completedWaves []*ActiveWave

	// Spawning logic
	formationFactory *entity.FormationFactory
	gameTime         float64
	lastWaveSpawn    float64
	minWaveInterval  float64 // minimum time between waves
	maxActiveWaves   int     // max simultaneous waves for performance

	// Level progression
	currentLevel   int
	wavesThisLevel int
	waveMultiplier float64 // increases difficulty

	// Boss management
	bossWaveActive bool
	bossThreshold  int // waves before boss
}

func NewWaveManager(eventManager interfaces.EventManagerInterface) *WaveManager {
	return &WaveManager{
		eventManager:     eventManager,
		eventChannels:    make(map[interfaces.EventType]<-chan interfaces.Event),
		pendingWaves:     make([]WaveDefinition, 0),
		activeWaves:      make([]*ActiveWave, 0),
		completedWaves:   make([]*ActiveWave, 0),
		formationFactory: entity.NewFormationFactory(eventManager),
		minWaveInterval:  config.Config.EnemySpawnInterval,
		maxActiveWaves:   4, // Allow 4 simultaneous waves for chaos
		currentLevel:     1,
		waveMultiplier:   1.0,
		bossThreshold:    config.Config.BossThreshold,
	}
}

func (wm *WaveManager) Initialize(ctx context.Context) error {
	err := wm.BaseSystem.Initialize(ctx)
	if err != nil {
		return err
	}

	// Subscribe to relevant events
	eventTypes := []interfaces.EventType{
		interfaces.SystemTick,
		interfaces.FormationDestroyed,
		interfaces.LevelChanged,
		interfaces.PlayerDestroyed,
	}

	for _, eventType := range eventTypes {
		ch, err := wm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event type %v: %w", eventType, err)
		}
		wm.eventChannels[eventType] = ch
	}

	// Initialize with starter waves for level 1
	wm.initializeLevel1Waves()

	return nil
}

func (wm *WaveManager) Update(deltaTime float64) error {
	select {
	case <-wm.CTX.Done():
		return wm.CTX.Err()
	default:
		wm.processAllAvailableEvents()
		wm.updateWaveSpawning()
		wm.updateActiveFormations()
		return nil
	}
}

func (wm *WaveManager) updateActiveFormations() {
	wm.mu.RLock()
	activeWaves := make([]*ActiveWave, len(wm.activeWaves))
	copy(activeWaves, wm.activeWaves)
	wm.mu.RUnlock()

	for _, wave := range activeWaves {
		if wave.Formation != nil && wave.State == WaveActive {
			wave.Formation.Update(core.FixedDeltaTime)
		}
	}
}

func (wm *WaveManager) eventListener() {
	systemTickCh := wm.eventChannels[interfaces.SystemTick]
	formationDestroyedCh := wm.eventChannels[interfaces.FormationDestroyed]
	levelChangedCh := wm.eventChannels[interfaces.LevelChanged]
	playerDestroyedCh := wm.eventChannels[interfaces.PlayerDestroyed]

	for {
		select {
		case <-wm.CTX.Done():
			return
		case _, ok := <-systemTickCh:
			if !ok {
				return
			}
			wm.processAllAvailableEvents()
			wm.updateWaveSpawning()
		case evt, ok := <-formationDestroyedCh:
			if !ok {
				return
			}
			wm.handleFormationDestroyed(evt)
		case evt, ok := <-levelChangedCh:
			if !ok {
				return
			}
			wm.handleLevelChanged(evt)
		case evt, ok := <-playerDestroyedCh:
			if !ok {
				return
			}
			wm.handlePlayerDestroyed(evt)
		}
	}
}

func (wm *WaveManager) processAllAvailableEvents() {
	if atomic.LoadInt32(&wm.isShutdown) == 1 {
		return
	}

	// Process any remaining events in channels
	eventChannels := []<-chan interfaces.Event{
		wm.eventChannels[interfaces.FormationDestroyed],
		wm.eventChannels[interfaces.LevelChanged],
		wm.eventChannels[interfaces.PlayerDestroyed],
	}

	handlers := []func(interfaces.Event){
		wm.handleFormationDestroyed,
		wm.handleLevelChanged,
		wm.handlePlayerDestroyed,
	}

	for i, ch := range eventChannels {
		for {
			select {
			case evt, ok := <-ch:
				if !ok {
					return
				}
				handlers[i](evt)
			default:
				goto nextChannel
			}
		}
	nextChannel:
	}
}

func (wm *WaveManager) updateWaveSpawning() {
	wm.mu.Lock()
	wm.gameTime += core.FixedDeltaTime
	currentTime := wm.gameTime
	wm.mu.Unlock()

	// UPDATE LES FORMATIONS ICI
	wm.updateActiveFormations()

	// Check if we can spawn new waves
	if wm.canSpawnWave(currentTime) {
		wm.spawnNextWave(currentTime)
	}

	// Update active wave states
	wm.updateActiveWaves()

	// Check level completion
	wm.checkLevelProgression()
}

func (wm *WaveManager) canSpawnWave(currentTime float64) bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	// No pending waves
	if len(wm.pendingWaves) == 0 {
		return false
	}

	// Too many active waves
	if len(wm.activeWaves) >= wm.maxActiveWaves {
		return false
	}

	// Too soon after last spawn
	if currentTime-wm.lastWaveSpawn < wm.minWaveInterval {
		return false
	}

	// Check if next wave is ready based on its delay
	nextWave := wm.pendingWaves[0]
	return currentTime >= wm.lastWaveSpawn+nextWave.SpawnDelay
}

func (wm *WaveManager) spawnNextWave(currentTime float64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if len(wm.pendingWaves) == 0 {
		return
	}

	// Get next wave to spawn
	waveDef := wm.pendingWaves[0]
	wm.pendingWaves = wm.pendingWaves[1:]

	// Create formation using factory
	formation := wm.formationFactory.CreatePresetFormation(
		waveDef.FormationType,
		waveDef.SpawnPosition,
		waveDef.EnemyType,
		waveDef.EnemyLevel,
		waveDef.EnemyCount,
	)

	// Create active wave
	activeWave := &ActiveWave{
		Definition: &waveDef,
		Formation:  formation,
		State:      WaveActive,
		SpawnTime:  currentTime,
	}

	wm.activeWaves = append(wm.activeWaves, activeWave)
	wm.lastWaveSpawn = currentTime
	wm.wavesThisLevel++

	// Publish wave started event
	wm.eventManager.Publish(interfaces.WaveStarted, map[string]interface{}{
		"wave_id":   waveDef.ID,
		"formation": formation,
		"level":     wm.currentLevel,
	})

	fmt.Printf("Spawned wave: %s (Active waves: %d)\n", waveDef.ID, len(wm.activeWaves))
}

func (wm *WaveManager) updateActiveWaves() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	activeWaves := make([]*ActiveWave, 0, len(wm.activeWaves))

	for _, wave := range wm.activeWaves {
		state := wave.Formation.GetState()

		if state == types.FormationDestroyed {
			// Move to completed
			wave.State = WaveCompleted
			wm.completedWaves = append(wm.completedWaves, wave)

			wm.eventManager.Publish(interfaces.WaveCompleted, map[string]interface{}{
				"wave_id":   wave.Definition.ID,
				"formation": wave.Formation,
			})
		} else {
			// Keep active
			activeWaves = append(activeWaves, wave)
		}
	}

	wm.activeWaves = activeWaves
}

func (wm *WaveManager) handleFormationDestroyed(evt interfaces.Event) {
	// Formation destruction is handled in updateActiveWaves
}

func (wm *WaveManager) handleLevelChanged(evt interfaces.Event) {
	if level, ok := evt.Data.(int); ok {
		wm.mu.Lock()
		wm.currentLevel = level
		wm.waveMultiplier += 0.2 // Increase difficulty
		wm.wavesThisLevel = 0
		wm.mu.Unlock()

		// Generate waves for new level
		wm.generateWavesForLevel(level)
	}
}

func (wm *WaveManager) handlePlayerDestroyed(evt interfaces.Event) {
	// Stop spawning new waves when player is destroyed
	wm.mu.Lock()
	wm.pendingWaves = wm.pendingWaves[:0]
	wm.mu.Unlock()
}

func (wm *WaveManager) checkLevelProgression() {
	wm.mu.RLock()
	wavesCompleted := len(wm.completedWaves)
	pendingCount := len(wm.pendingWaves)
	activeCount := len(wm.activeWaves)
	shouldSpawnBoss := wavesCompleted >= wm.bossThreshold && !wm.bossWaveActive
	wm.mu.RUnlock()

	// Level complete when all waves done and no boss active
	if pendingCount == 0 && activeCount == 0 && !wm.bossWaveActive {
		// Advance to next level
		wm.eventManager.Publish(interfaces.LevelChanged, wm.currentLevel+1)
	} else if shouldSpawnBoss {
		wm.spawnBossWave()
	}
}

func (wm *WaveManager) spawnBossWave() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.bossWaveActive {
		return
	}

	wm.bossWaveActive = true

	// Create boss at top center of screen
	bossPos := types.Vector2D{
		X: float64(config.Config.ScreenWidth / 2),
		Y: 50,
	}

	boss := entity.NewBoss(bossPos, wm.eventManager)

	// Notify that boss spawned
	wm.eventManager.Publish(interfaces.EnemyCreated, boss)

	fmt.Printf("Boss spawned at level %d!\n", wm.currentLevel)
}

// WAVE GENERATION - Creates dynamic wave patterns

func (wm *WaveManager) initializeLevel1Waves() {
	waves := []WaveDefinition{
		{
			ID:            "level1_wave1_scouts",
			FormationType: entity.PresetVFormation,
			EnemyType:     graphics.Scout,
			EnemyLevel:    graphics.Level1,
			EnemyCount:    5,
			SpawnDelay:    0.0,
			SpawnPosition: types.Vector2D{X: 320, Y: -50},
			Priority:      1,
		},
		{
			ID:            "level1_wave2_line",
			FormationType: entity.PresetLineFormation,
			EnemyType:     graphics.Scout,
			EnemyLevel:    graphics.Level1,
			EnemyCount:    14,
			SpawnDelay:    3.0,
			SpawnPosition: types.Vector2D{X: 200, Y: -50},
			Priority:      2,
		},
		{
			ID:            "level1_wave3_fighters",
			FormationType: entity.PresetSineWaveFormation,
			EnemyType:     graphics.Fighter,
			EnemyLevel:    graphics.Level1,
			EnemyCount:    6,
			SpawnDelay:    2.5,
			SpawnPosition: types.Vector2D{X: 450, Y: -50},
			Priority:      3,
		},
	}

	wm.mu.Lock()
	wm.pendingWaves = append(wm.pendingWaves, waves...)
	wm.mu.Unlock()
}

func (wm *WaveManager) generateWavesForLevel(level int) {
	// Generate dynamic waves based on level
	baseEnemyCount := 3 + level
	waveCount := 4 + (level * 2)

	waves := make([]WaveDefinition, 0, waveCount)

	for i := 0; i < waveCount; i++ {
		// Randomize formation types and positions for variety
		formationType := entity.PresetFormationType(rand.Intn(4))
		enemyType := graphics.EnemyType(rand.Intn(3)) // Scout, Fighter, Heavy

		spawnX := 100 + rand.Float64()*440 // Random X between 100-540

		wave := WaveDefinition{
			ID:            fmt.Sprintf("level%d_wave%d", level, i+1),
			FormationType: formationType,
			EnemyType:     enemyType,
			EnemyLevel:    graphics.EnemyLevel(min(level-1, 4)), // Cap at Level5
			EnemyCount:    baseEnemyCount + rand.Intn(3),
			SpawnDelay:    2.0 + rand.Float64()*3.0, // 2-5 seconds
			SpawnPosition: types.Vector2D{X: spawnX, Y: -50},
			Priority:      i + 1,
		}

		waves = append(waves, wave)
	}

	wm.mu.Lock()
	wm.pendingWaves = append(wm.pendingWaves, waves...)
	wm.completedWaves = wm.completedWaves[:0] // Reset completed waves
	wm.mu.Unlock()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Public API for external control

func (wm *WaveManager) GetActiveWaveCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.activeWaves)
}

func (wm *WaveManager) GetPendingWaveCount() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return len(wm.pendingWaves)
}

func (wm *WaveManager) GetCurrentLevel() int {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.currentLevel
}

func (wm *WaveManager) ForceSpawnWave() {
	wm.mu.Lock()
	currentTime := wm.gameTime
	wm.mu.Unlock()

	if wm.canSpawnWave(currentTime) {
		wm.spawnNextWave(currentTime)
	}
}

func (wm *WaveManager) Shutdown() {
	if !atomic.CompareAndSwapInt32(&wm.isShutdown, 0, 1) {
		return
	}

	wm.mu.Lock()
	defer wm.mu.Unlock()

	for eventType, ch := range wm.eventChannels {
		wm.eventManager.Unsubscribe(eventType, ch)
	}
	wm.eventChannels = nil

	wm.pendingWaves = nil
	wm.activeWaves = nil
	wm.completedWaves = nil
}

var _ core.System = (*WaveManager)(nil)
