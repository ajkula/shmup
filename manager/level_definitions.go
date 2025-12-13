package manager

type LevelDefinition struct {
	Name  string       `json:"name"`
	Waves []WaveConfig `json:"waves"`
	Boss  BossConfig   `json:"boss"`
}

type WaveConfig struct {
	ID         string   `json:"id"`
	Time       float64  `json:"time"`
	Formation  string   `json:"formation"`
	EnemyType  string   `json:"enemyType"`
	EnemyLevel int      `json:"enemyLevel"`
	EnemyCount int      `json:"enemyCount"`
	Position   Position `json:"position"`
	Speed      float64  `json:"speed"`
	Spacing    float64  `json:"spacing,omitempty"`
	Radius     float64  `json:"radius,omitempty"`
	Amplitude  float64  `json:"amplitude,omitempty"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type BossConfig struct {
	AppearsAfterWaves int      `json:"appearsAfterWaves"`
	Type              string   `json:"type"`
	Health            int      `json:"health"`
	Position          Position `json:"position"`
}

// Level 1: Introduction - Classic formations
var Level1Definition = LevelDefinition{
	Name: "LEVEL 1: FIRST CONTACT",
	Waves: []WaveConfig{
		// Opening: Classic V formation
		{
			ID:         "wave1",
			Time:       0.5,
			Formation:  "V",
			EnemyType:  "Scout",
			EnemyLevel: 1,
			EnemyCount: 7,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.2,
			Spacing:    48,
		},
		// Arrow attack from above
		{
			ID:         "wave2",
			Time:       3.5,
			Formation:  "Arrow",
			EnemyType:  "Scout",
			EnemyLevel: 1,
			EnemyCount: 9,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.5,
			Spacing:    42,
		},
		// Wing formation from sides
		{
			ID:         "wave3",
			Time:       3.0,
			Formation:  "Wings",
			EnemyType:  "Fighter",
			EnemyLevel: 1,
			EnemyCount: 8,
			Position:   Position{X: 320, Y: -50},
			Speed:      1.9,
			Spacing:    38,
			Radius:     130,
		},
		// Diamond formation
		{
			ID:         "wave4",
			Time:       3.5,
			Formation:  "Diamond",
			EnemyType:  "Scout",
			EnemyLevel: 1,
			EnemyCount: 9,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.3,
			Spacing:    45,
		},
		// Line of fighters
		{
			ID:         "wave5",
			Time:       3.0,
			Formation:  "Line",
			EnemyType:  "Fighter",
			EnemyLevel: 1,
			EnemyCount: 5,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.4,
			Spacing:    55,
		},
	},
	Boss: BossConfig{
		AppearsAfterWaves: 5,
		Type:              "StandardBoss",
		Health:            500,
		Position:          Position{X: -1, Y: 100},
	},
}

// Level 2: Escalation - Advanced patterns
var Level2Definition = LevelDefinition{
	Name: "LEVEL 2: RISING THREAT",
	Waves: []WaveConfig{
		// Spiral entrance
		{
			ID:         "wave1",
			Time:       0.5,
			Formation:  "Spiral",
			EnemyType:  "Fighter",
			EnemyLevel: 1,
			EnemyCount: 12,
			Position:   Position{X: 320, Y: -50},
			Speed:      1.8,
			Spacing:    12,
			Radius:     25,
		},
		// ZigZag attack
		{
			ID:         "wave2",
			Time:       3.0,
			Formation:  "ZigZag",
			EnemyType:  "Scout",
			EnemyLevel: 2,
			EnemyCount: 8,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.6,
			Amplitude:  90,
			Spacing:    48,
		},
		// Rotating circle
		{
			ID:         "wave3",
			Time:       3.5,
			Formation:  "Circle",
			EnemyType:  "Fighter",
			EnemyLevel: 1,
			EnemyCount: 10,
			Position:   Position{X: 200, Y: -50},
			Speed:      1.7,
			Radius:     95,
		},
		// Double V attack
		{
			ID:         "wave4",
			Time:       2.5,
			Formation:  "V",
			EnemyType:  "Heavy",
			EnemyLevel: 1,
			EnemyCount: 7,
			Position:   Position{X: 200, Y: -50},
			Speed:      2.0,
			Spacing:    50,
		},
		{
			ID:         "wave5",
			Time:       0.5,
			Formation:  "V",
			EnemyType:  "Heavy",
			EnemyLevel: 1,
			EnemyCount: 7,
			Position:   Position{X: 440, Y: -50},
			Speed:      2.0,
			Spacing:    50,
		},
		// Sine wave finale
		{
			ID:         "wave6",
			Time:       3.5,
			Formation:  "SineWave",
			EnemyType:  "Fighter",
			EnemyLevel: 2,
			EnemyCount: 10,
			Position:   Position{X: 320, Y: -50},
			Speed:      1.8,
			Amplitude:  110,
			Spacing:    48,
		},
	},
	Boss: BossConfig{
		AppearsAfterWaves: 6,
		Type:              "StandardBoss",
		Health:            700,
		Position:          Position{X: -1, Y: 100},
	},
}

// Level 3: Chaos - Mix of everything
var Level3Definition = LevelDefinition{
	Name: "LEVEL 3: TOTAL CHAOS",
	Waves: []WaveConfig{
		// Arrows from both sides
		{
			ID:         "wave1",
			Time:       0.5,
			Formation:  "Arrow",
			EnemyType:  "Fighter",
			EnemyLevel: 2,
			EnemyCount: 9,
			Position:   Position{X: 180, Y: -50},
			Speed:      2.7,
			Spacing:    40,
		},
		{
			ID:         "wave2",
			Time:       0.2,
			Formation:  "Arrow",
			EnemyType:  "Fighter",
			EnemyLevel: 2,
			EnemyCount: 9,
			Position:   Position{X: 460, Y: -50},
			Speed:      2.7,
			Spacing:    40,
		},
		// Spiral madness
		{
			ID:         "wave3",
			Time:       3.5,
			Formation:  "Spiral",
			EnemyType:  "Heavy",
			EnemyLevel: 1,
			EnemyCount: 14,
			Position:   Position{X: 320, Y: -50},
			Speed:      1.6,
			Spacing:    14,
			Radius:     20,
		},
		// Wings assault
		{
			ID:         "wave4",
			Time:       3.0,
			Formation:  "Wings",
			EnemyType:  "Scout",
			EnemyLevel: 2,
			EnemyCount: 10,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.5,
			Spacing:    42,
			Radius:     140,
		},
		// Diamond formation
		{
			ID:         "wave5",
			Time:       3.0,
			Formation:  "Diamond",
			EnemyType:  "Heavy",
			EnemyLevel: 2,
			EnemyCount: 13,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.1,
			Spacing:    48,
		},
		// ZigZag chaos
		{
			ID:         "wave6",
			Time:       3.0,
			Formation:  "ZigZag",
			EnemyType:  "Fighter",
			EnemyLevel: 2,
			EnemyCount: 10,
			Position:   Position{X: 320, Y: -50},
			Speed:      2.8,
			Amplitude:  100,
			Spacing:    50,
		},
		// Circle finale
		{
			ID:         "wave7",
			Time:       3.5,
			Formation:  "Circle",
			EnemyType:  "Heavy",
			EnemyLevel: 2,
			EnemyCount: 12,
			Position:   Position{X: 320, Y: -50},
			Speed:      1.5,
			Radius:     100,
		},
	},
	Boss: BossConfig{
		AppearsAfterWaves: 7,
		Type:              "StandardBoss",
		Health:            900,
		Position:          Position{X: -1, Y: 100},
	},
}

var AllLevels = []LevelDefinition{
	Level1Definition,
	Level2Definition,
	Level3Definition,
}

func GetLevelDefinition(levelNum int) *LevelDefinition {
	if levelNum > 0 && levelNum <= len(AllLevels) {
		return &AllLevels[levelNum-1]
	}
	return &Level1Definition
}
