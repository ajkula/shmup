package config

import (
	"os"
	"strconv"
)

type GameConfig struct {
	ScreenWidth  int
	ScreenHeight int
	PlayerSpeed  float64
	EnemySpeed   float64
	BulletSpeed  float64

	BossThreshold      int
	EnemySpawnInterval float64
	PowerUpSpawnChance float64

	MaxEventQueueSize int
	MaxStateQueueSize int
}

var Config GameConfig

func Init() {
	Config = GameConfig{
		ScreenWidth:        640,
		ScreenHeight:       928,
		PlayerSpeed:        5.0,
		EnemySpeed:         2.0,
		BulletSpeed:        10.0,
		BossThreshold:      50,
		EnemySpawnInterval: 2.0,
		PowerUpSpawnChance: 0.1,
		MaxEventQueueSize:  100,
		MaxStateQueueSize:  10,
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
