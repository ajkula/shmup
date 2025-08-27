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

var Level1Definition = LevelDefinition{
	Name: "Level 1",
	Waves: []WaveConfig{
		{
			ID:         "wave1",
			Time:       1.0,
			Formation:  "V",
			EnemyType:  "Scout",
			EnemyLevel: 1,
			EnemyCount: 9,
			Position:   Position{X: 320, Y: -100},
			Speed:      2.0,
			Spacing:    50,
		},
		{
			ID:         "wave2",
			Time:       1.5,
			Formation:  "Line",
			EnemyType:  "Scout",
			EnemyLevel: 1,
			EnemyCount: 10,
			Position:   Position{X: 320, Y: -80},
			Speed:      2.5,
			Spacing:    60,
		},
		{
			ID:         "wave3",
			Time:       2.0,
			Formation:  "Circle",
			EnemyType:  "Fighter",
			EnemyLevel: 1,
			EnemyCount: 18,
			Position:   Position{X: 200, Y: -120},
			Speed:      1.8,
			Radius:     110,
		},
		{
			ID:         "wave4",
			Time:       4.0,
			Formation:  "SineWave",
			EnemyType:  "Heavy",
			EnemyLevel: 1,
			EnemyCount: 18,
			Position:   Position{X: 320, Y: -100},
			Speed:      1.5,
			Amplitude:  100,
			Spacing:    45,
		},
	},
	Boss: BossConfig{
		AppearsAfterWaves: 10,
		Type:              "StandardBoss",
		Health:            1000,
		Position:          Position{X: 320, Y: 100},
	},
}

var AllLevels = []LevelDefinition{
	Level1Definition,
	// Level2Definition, ...
}

func GetLevelDefinition(levelNum int) *LevelDefinition {
	if levelNum > 0 && levelNum <= len(AllLevels) {
		return &AllLevels[levelNum-1]
	}
	return &Level1Definition
}
