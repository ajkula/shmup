package config

import (
	"image/color"
)

type GameConfig struct {
	ScreenWidth  int
	ScreenHeight int
	PlayerSpeed  float64
	EnemySpeed   float64
	BulletSpeed  float64

	BossThreshold      int
	EnemySpawnInterval float64

	BackgroundColor color.Color
}

var Config GameConfig

func Init() {
	Config = GameConfig{
		ScreenWidth:        640,
		ScreenHeight:       928,
		PlayerSpeed:        5.0,
		EnemySpeed:         2.0,
		BulletSpeed:        10.0,
		BossThreshold:      5, // waves before BOSS
		EnemySpawnInterval: 2.0,
		BackgroundColor:    color.RGBA{0, 0, 0, 255},
	}
}
