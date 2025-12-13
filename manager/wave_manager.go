package manager

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/entity"
	"github.com/ajkula/shmup/graphics"
	"github.com/ajkula/shmup/interfaces"
	"github.com/ajkula/shmup/types"
)

type WaveDefinition struct {
	ID            string
	FormationType entity.PresetFormationType
	EnemyType     graphics.EnemyType
	EnemyLevel    graphics.EnemyLevel
	EnemyCount    int
	SpawnDelay    float64 // seconds after previous wave
	SpawnPosition types.Vector2D
	Priority      int // higher = spawns first if multiple ready
	PatternConfig *entity.PatternConfig
}

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
	minWaveInterval  float64
	maxActiveWaves   int

	// Level progression
	currentLevel   int
	wavesThisLevel int
	waveMultiplier float64

	// Boss management
	bossWaveActive        bool
	bossDefeatedThisLevel bool
	bossThreshold         int
	waitingForVictory     bool // Flag to indicate we're waiting for all waves to clear before Victory
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
		interfaces.BossDefeated,
	}

	for _, eventType := range eventTypes {
		ch, err := wm.eventManager.Subscribe(eventType)
		if err != nil {
			return fmt.Errorf("failed to subscribe to event type %v: %w", eventType, err)
		}
		wm.eventChannels[eventType] = ch
	}

	wm.generateWavesForLevel(1)

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

func (wm *WaveManager) processAllAvailableEvents() {
	if atomic.LoadInt32(&wm.isShutdown) == 1 {
		return
	}

	// Process any remaining events in channels
	eventChannels := []<-chan interfaces.Event{
		wm.eventChannels[interfaces.FormationDestroyed],
		wm.eventChannels[interfaces.LevelChanged],
		wm.eventChannels[interfaces.PlayerDestroyed],
		wm.eventChannels[interfaces.BossDefeated],
	}

	handlers := []func(interfaces.Event){
		wm.handleFormationDestroyed,
		wm.handleLevelChanged,
		wm.handlePlayerDestroyed,
		wm.handleBossDefeated,
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

	// Don't spawn waves if boss is active or defeated
	if wm.bossWaveActive || wm.bossDefeatedThisLevel {
		return false
	}

	// Too many active waves
	if len(wm.activeWaves) >= wm.maxActiveWaves {
		fmt.Printf("Cannot spawn: too many active waves (%d/%d)\n", len(wm.activeWaves), wm.maxActiveWaves)
		return false
	}

	// Too soon after last spawn
	timeSinceLastSpawn := currentTime - wm.lastWaveSpawn
	if timeSinceLastSpawn < wm.minWaveInterval {
		return false
	}

	// Check if next wave is ready based on its delay
	nextWave := wm.pendingWaves[0]
	readyTime := wm.lastWaveSpawn + nextWave.SpawnDelay
	canSpawn := currentTime >= readyTime

	if !canSpawn && int(currentTime)%60 == 0 {
		fmt.Printf("Wave not ready: current=%.2f, lastSpawn=%.2f, delay=%.2f, ready=%.2f\n",
			currentTime, wm.lastWaveSpawn, nextWave.SpawnDelay, readyTime)
	}

	return canSpawn
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
	fmt.Printf("Wave spawned, remaining pending: %d\n", len(wm.pendingWaves))

	var formation types.FormationController

	if waveDef.PatternConfig != nil {
		// JSON
		formation = wm.formationFactory.CreateFormationWithConfig(
			waveDef.FormationType,
			waveDef.SpawnPosition,
			waveDef.EnemyType,
			waveDef.EnemyLevel,
			waveDef.EnemyCount,
			*waveDef.PatternConfig,
		)
	} else {
		// defaut
		formation = wm.formationFactory.CreatePresetFormation(
			waveDef.FormationType,
			waveDef.SpawnPosition,
			waveDef.EnemyType,
			waveDef.EnemyLevel,
			waveDef.EnemyCount,
		)
	}

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
	wm.eventManager.Publish(interfaces.WaveStarted, map[string]any{
		"wave_id":   waveDef.ID,
		"formation": formation,
		"level":     wm.currentLevel,
	})
	time := time.Now()
	fmt.Printf("%s Spawned wave: %s (Active waves: %d)\n", time, waveDef.ID, len(wm.activeWaves))
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

			fmt.Printf("Wave completed: %s (Active: %d -> %d, Completed: %d)\n",
				wave.Definition.ID, len(wm.activeWaves), len(wm.activeWaves)-1, len(wm.completedWaves))

			wm.eventManager.Publish(interfaces.WaveCompleted, map[string]any{
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
	// noop
}

func (wm *WaveManager) handleLevelChanged(evt interfaces.Event) {
	if level, ok := evt.Data.(int); ok {
		wm.mu.Lock()
		wm.currentLevel = level
		wm.waveMultiplier += 0.2
		wm.wavesThisLevel = 0
		wm.mu.Unlock()

		wm.generateWavesForLevel(level)
	}
}

func (wm *WaveManager) handlePlayerDestroyed(evt interfaces.Event) {
	wm.mu.Lock()
	wm.pendingWaves = wm.pendingWaves[:0]
	wm.mu.Unlock()
}

func (wm *WaveManager) checkLevelProgression() {
	wm.mu.RLock()

	wavesCompletedThisLevel := 0
	levelPrefix := fmt.Sprintf("level%d_", wm.currentLevel)
	for _, wave := range wm.completedWaves {
		if wave.Definition != nil && strings.HasPrefix(wave.Definition.ID, levelPrefix) {
			wavesCompletedThisLevel++
		}
	}

	pendingCount := len(wm.pendingWaves)
	activeCount := len(wm.activeWaves)
	currentLevel := wm.currentLevel
	bossActive := wm.bossWaveActive
	bossDefeated := wm.bossDefeatedThisLevel
	threshold := wm.bossThreshold
	wm.mu.RUnlock()

	if wavesCompletedThisLevel >= threshold && !bossActive && !bossDefeated && activeCount == 0 && pendingCount == 0 {
		fmt.Printf("Boss conditions met! Spawning boss...\n")
		wm.spawnBossWave()
		return
	}

	if bossActive {
		return
	}

	if bossDefeated && pendingCount == 0 && activeCount == 0 {
		// Check if there's a next level
		nextLevel := currentLevel + 1
		hasNextLevel := nextLevel <= len(AllLevels)

		if hasNextLevel {
			// Continue to next level
			wm.mu.Lock()
			wm.currentLevel++
			wm.wavesThisLevel = 0
			wm.completedWaves = wm.completedWaves[:0]
			wm.bossDefeatedThisLevel = false
			wm.mu.Unlock()

			fmt.Printf("Level %d complete! Starting level %d\n", currentLevel, wm.currentLevel)
			wm.generateWavesForLevel(wm.currentLevel)
		} else {
			// No more levels - trigger Victory
			fmt.Println("All levels completed - VICTORY!")
			wm.mu.Lock()
			wm.waitingForVictory = false // Reset flag
			wm.mu.Unlock()
			wm.eventManager.Publish(interfaces.Victory, nil)
		}
	}
}

func (wm *WaveManager) spawnBossWave() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.bossWaveActive {
		return
	}

	wm.bossWaveActive = true

	levelDef := GetLevelDefinition(wm.currentLevel)
	bossConfig := levelDef.Boss

	bossPos := types.Vector2D{
		X: bossConfig.Position.X,
		Y: bossConfig.Position.Y,
	}

	if bossConfig.Position.X < 0 {
		bossPos.X = float64(config.Config.ScreenWidth / 2)
	}

	boss := entity.NewBoss(bossPos, wm.eventManager)
	boss.Health = bossConfig.Health
	boss.MaxHealth = bossConfig.Health

	wm.eventManager.Publish(interfaces.EnemyCreated, boss)

	fmt.Printf("Boss spawned at level %d!\n", wm.currentLevel)
}

func (wm *WaveManager) handleBossDefeated(evt interfaces.Event) {
	wm.mu.Lock()
	wm.bossWaveActive = false
	currentLevel := wm.currentLevel
	wm.bossDefeatedThisLevel = true
	// Clear pending waves immediately to prevent any spawns during transition
	wm.pendingWaves = wm.pendingWaves[:0]
	wm.mu.Unlock()

	fmt.Printf("%s Boss defeated at level %d!\n", time.Now(), currentLevel)

	// checkLevelProgression() will handle level transition or Victory
	// once all active waves are cleared
}

// WAVE GENERATION

func (wm *WaveManager) generateWavesForLevel(level int) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.generateWavesForLevelUnsafe(level)
}

func (wm *WaveManager) generateWavesForLevelUnsafe(level int) {
	levelDef := GetLevelDefinition(level)

	wm.pendingWaves = make([]WaveDefinition, 0)

	for _, waveConfig := range levelDef.Waves {
		wave := WaveDefinition{
			ID:            fmt.Sprintf("level%d_%s", level, waveConfig.ID),
			FormationType: parseFormationType(waveConfig.Formation),
			EnemyType:     parseEnemyType(waveConfig.EnemyType),
			EnemyLevel:    graphics.EnemyLevel(waveConfig.EnemyLevel - 1 + level),
			EnemyCount:    waveConfig.EnemyCount,
			SpawnDelay:    waveConfig.Time,
			SpawnPosition: types.Vector2D{X: waveConfig.Position.X, Y: waveConfig.Position.Y},
			Priority:      1,
		}

		if waveConfig.Speed > 0 || waveConfig.Spacing > 0 {
			wave.PatternConfig = &entity.PatternConfig{
				Speed:         60.0,
				VerticalSpeed: waveConfig.Speed * 60.0,
				Spacing:       waveConfig.Spacing,
				Radius:        waveConfig.Radius,
				Amplitude:     waveConfig.Amplitude,
				Frequency:     2.0,
				RotationSpeed: 1.0,
			}
		}

		wm.pendingWaves = append(wm.pendingWaves, wave)
	}

	wm.bossThreshold = levelDef.Boss.AppearsAfterWaves
}

// Fonctions helper
func parseFormationType(formation string) entity.PresetFormationType {
	switch formation {
	case "V":
		return entity.PresetVFormation
	case "Line":
		return entity.PresetLineFormation
	case "Circle":
		return entity.PresetCircleFormation
	case "SineWave":
		return entity.PresetSineWaveFormation
	case "Diamond":
		return entity.PresetDiamondFormation
	case "Wings":
		return entity.PresetWingsFormation
	case "Spiral":
		return entity.PresetSpiralFormation
	case "Arrow":
		return entity.PresetArrowFormation
	case "ZigZag":
		return entity.PresetZigZagFormation
	default:
		return entity.PresetVFormation
	}
}

func parseEnemyType(enemyType string) graphics.EnemyType {
	switch enemyType {
	case "Scout":
		return graphics.Scout
	case "Fighter":
		return graphics.Fighter
	case "Heavy":
		return graphics.Heavy
	default:
		return graphics.Scout
	}
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

func (wm *WaveManager) ResetToLevel1() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	fmt.Println("[WaveManager] Resetting to level 1")

	// Clear all waves
	wm.pendingWaves = wm.pendingWaves[:0]
	wm.activeWaves = wm.activeWaves[:0]
	wm.completedWaves = wm.completedWaves[:0]

	// Reset level state
	wm.currentLevel = 1
	wm.wavesThisLevel = 0
	wm.waveMultiplier = 1.0
	wm.gameTime = 0
	wm.lastWaveSpawn = 0

	// Reset boss state
	wm.bossWaveActive = false
	wm.bossDefeatedThisLevel = false
	wm.waitingForVictory = false

	// Generate level 1 waves (use Unsafe version since we already have the lock)
	wm.generateWavesForLevelUnsafe(1)

	fmt.Printf("[WaveManager] Reset complete - %d waves ready for level 1\n", len(wm.pendingWaves))
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
