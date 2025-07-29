package graphics

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type EnemyType int

const (
	Scout EnemyType = iota
	Fighter
	Heavy
	Boss
)

type EnemyLevel int

const (
	Level1 EnemyLevel = iota
	Level2
	Level3
	Level4
	Level5
)

type PlayerClass int

const (
	StandardFighter PlayerClass = iota
	HeavyFighter
	Interceptor
)

type SpriteData struct {
	Pixels [][]color.Color // Pure pixel grid
	Width  int
	Height int
	Image  *ebiten.Image // Pre-rendered image for performance
}

type EnemyStats struct {
	HP       int
	Speed    float64
	Damage   int
	FireRate float64
}

var (
	// Colors
	transparent = color.RGBA{0, 0, 0, 0}
	gray        = color.RGBA{128, 128, 128, 255}
	lightGray   = color.RGBA{180, 180, 180, 255}
	darkGray    = color.RGBA{80, 80, 80, 255}
	cyan        = color.RGBA{0, 255, 255, 255}
	blue        = color.RGBA{0, 128, 255, 255}
	orange      = color.RGBA{255, 128, 0, 255}
	red         = color.RGBA{255, 0, 0, 255}
	green       = color.RGBA{0, 255, 0, 255}
	darkGreen   = color.RGBA{0, 150, 0, 255}
	yellow      = color.RGBA{255, 255, 0, 255}
	magenta     = color.RGBA{255, 0, 255, 255}
	darkBlue    = color.RGBA{0, 0, 150, 255}
)

var playerSprites = map[PlayerClass]*SpriteData{
	StandardFighter: {
		Pixels: [][]color.Color{
			{transparent, transparent, transparent, cyan, transparent, transparent, transparent},
			{transparent, transparent, cyan, cyan, cyan, transparent, transparent},
			{transparent, cyan, gray, gray, gray, cyan, transparent},
			{cyan, gray, gray, gray, gray, gray, cyan},
			{cyan, gray, gray, blue, gray, gray, cyan},
			{gray, gray, gray, gray, gray, gray, gray},
			{transparent, gray, gray, gray, gray, gray, transparent},
			{transparent, transparent, gray, orange, gray, transparent, transparent},
			{transparent, transparent, gray, orange, gray, transparent, transparent},
		},
		Width: 7, Height: 9,
	},
}

var enemySprites = map[EnemyType]map[EnemyLevel]*SpriteData{
	Scout: {
		Level1: {
			Pixels: [][]color.Color{
				{transparent, transparent, gray, gray, gray, transparent, transparent},
				{transparent, gray, darkGray, darkGray, darkGray, gray, transparent},
				{gray, darkGray, darkGray, red, darkGray, darkGray, gray},
				{gray, darkGray, darkGray, green, darkGray, darkGray, gray},
				{transparent, gray, darkGray, darkGray, darkGray, gray, transparent},
				{transparent, transparent, gray, gray, gray, transparent, transparent},
				{transparent, transparent, transparent, green, transparent, transparent, transparent},
			},
			Width: 7, Height: 7,
		},
		Level2: {
			Pixels: [][]color.Color{
				{transparent, transparent, darkGreen, darkGreen, darkGreen, transparent, transparent},
				{transparent, darkGreen, green, green, green, darkGreen, transparent},
				{darkGreen, green, green, yellow, green, green, darkGreen},
				{darkGreen, green, green, cyan, green, green, darkGreen},
				{transparent, darkGreen, green, green, green, darkGreen, transparent},
				{transparent, transparent, darkGreen, darkGreen, darkGreen, transparent, transparent},
				{transparent, transparent, cyan, cyan, cyan, transparent, transparent},
			},
			Width: 7, Height: 7,
		},
	},
	Fighter: {
		Level1: {
			Pixels: [][]color.Color{
				{transparent, transparent, transparent, lightGray, transparent, transparent, transparent},
				{transparent, transparent, lightGray, lightGray, lightGray, transparent, transparent},
				{transparent, lightGray, gray, gray, gray, lightGray, transparent},
				{lightGray, gray, gray, gray, gray, gray, lightGray},
				{gray, gray, gray, gray, gray, gray, gray},
				{transparent, gray, gray, gray, gray, gray, transparent},
				{transparent, transparent, gray, gray, gray, transparent, transparent},
				{transparent, transparent, transparent, orange, transparent, transparent, transparent},
			},
			Width: 7, Height: 8,
		},
	},
	Heavy: {
		Level1: {
			Pixels: [][]color.Color{
				{lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray},
				{lightGray, gray, gray, gray, gray, gray, lightGray},
				{lightGray, gray, lightGray, lightGray, lightGray, gray, lightGray},
				{lightGray, gray, lightGray, red, lightGray, gray, lightGray},
				{lightGray, gray, lightGray, lightGray, lightGray, gray, lightGray},
				{lightGray, gray, gray, gray, gray, gray, lightGray},
				{lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray},
				{transparent, orange, orange, orange, orange, orange, transparent},
			},
			Width: 7, Height: 8,
		},
	},
	Boss: {
		Level1: {
			Pixels: [][]color.Color{
				{transparent, transparent, transparent, magenta, magenta, transparent, transparent, transparent, transparent, magenta, magenta, transparent, transparent, transparent},
				{transparent, transparent, darkBlue, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkBlue, transparent, transparent},
				{transparent, darkGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGray, transparent},
				{darkGray, lightGray, lightGray, cyan, lightGray, blue, lightGray, blue, lightGray, cyan, lightGray, lightGray, lightGray, darkGray},
				{darkGray, lightGray, lightGray, lightGray, lightGray, lightGray, red, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGray},
				{darkGray, lightGray, cyan, lightGray, magenta, lightGray, lightGray, lightGray, magenta, lightGray, cyan, lightGray, lightGray, darkGray},
				{darkGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGray},
				{transparent, darkGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGray, transparent},
				{transparent, transparent, darkGray, darkGray, darkGray, darkGray, darkGray, darkGray, darkGray, darkGray, darkGray, darkGray, transparent, transparent},
				{transparent, transparent, transparent, orange, transparent, orange, transparent, orange, transparent, orange, transparent, transparent, transparent, transparent},
				{transparent, transparent, transparent, orange, transparent, orange, transparent, orange, transparent, orange, transparent, transparent, transparent, transparent},
			},
			Width: 14, Height: 11,
		},
		Level2: {
			Pixels: [][]color.Color{
				{transparent, magenta, transparent, transparent, magenta, magenta, transparent, transparent, magenta, magenta, transparent, transparent, magenta, transparent},
				{transparent, darkBlue, darkBlue, darkGreen, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGreen, darkBlue, darkBlue, transparent},
				{darkBlue, darkGreen, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGreen, darkBlue},
				{darkGreen, lightGray, lightGray, cyan, lightGray, yellow, lightGray, yellow, lightGray, cyan, lightGray, lightGray, lightGray, darkGreen},
				{darkGreen, lightGray, lightGray, lightGray, lightGray, lightGray, magenta, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGreen},
				{darkGreen, lightGray, cyan, lightGray, magenta, lightGray, red, lightGray, magenta, lightGray, cyan, lightGray, lightGray, darkGreen},
				{darkGreen, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGreen},
				{darkGreen, orange, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, orange, darkGreen},
				{darkBlue, darkGreen, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkGreen, darkBlue},
				{transparent, darkBlue, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkGreen, darkBlue, transparent},
				{transparent, transparent, orange, transparent, orange, transparent, orange, orange, transparent, orange, transparent, orange, transparent, transparent},
				{transparent, transparent, orange, transparent, orange, transparent, orange, orange, transparent, orange, transparent, orange, transparent, transparent},
			},
			Width: 14, Height: 12,
		},
		Level3: {
			Pixels: [][]color.Color{
				{magenta, transparent, magenta, transparent, transparent, magenta, magenta, magenta, magenta, transparent, transparent, magenta, transparent, magenta},
				{darkBlue, darkBlue, darkBlue, darkBlue, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkBlue, darkBlue, darkBlue, darkBlue},
				{darkBlue, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkBlue},
				{darkBlue, lightGray, cyan, cyan, lightGray, yellow, yellow, yellow, yellow, lightGray, cyan, cyan, lightGray, darkBlue},
				{darkBlue, lightGray, cyan, lightGray, lightGray, lightGray, magenta, magenta, lightGray, lightGray, lightGray, cyan, lightGray, darkBlue},
				{darkBlue, lightGray, lightGray, lightGray, magenta, magenta, red, red, magenta, magenta, lightGray, lightGray, lightGray, darkBlue},
				{darkBlue, lightGray, lightGray, lightGray, lightGray, lightGray, red, red, lightGray, lightGray, lightGray, lightGray, lightGray, darkBlue},
				{darkBlue, orange, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, orange, darkBlue},
				{darkBlue, orange, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, orange, darkBlue},
				{darkBlue, darkBlue, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, lightGray, darkBlue, darkBlue},
				{transparent, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, darkBlue, transparent},
				{transparent, orange, transparent, orange, transparent, orange, orange, orange, orange, transparent, orange, transparent, orange, transparent},
				{transparent, orange, transparent, orange, transparent, orange, orange, orange, orange, transparent, orange, transparent, orange, transparent},
			},
			Width: 14, Height: 13,
		},
	},
}

var enemyStats = map[EnemyType]map[EnemyLevel]EnemyStats{
	Scout: {
		Level1: {HP: 1, Speed: 3.0, Damage: 5, FireRate: 2.0},
		Level2: {HP: 2, Speed: 4.0, Damage: 8, FireRate: 1.8},
	},
	Fighter: {
		Level1: {HP: 5, Speed: 2.0, Damage: 10, FireRate: 1.5},
	},
	Heavy: {
		Level1: {HP: 15, Speed: 1.0, Damage: 20, FireRate: 3.0},
	},
}

func createSpriteImage(sprite *SpriteData) *ebiten.Image {
	pixelSize := 4
	img := ebiten.NewImage(sprite.Width*pixelSize, sprite.Height*pixelSize)

	for row := 0; row < sprite.Height; row++ {
		for col := 0; col < sprite.Width; col++ {
			pixelColor := sprite.Pixels[row][col]
			if pixelColor == transparent {
				continue
			}

			// Draw 4x4 pixel block
			for py := 0; py < pixelSize; py++ {
				for px := 0; px < pixelSize; px++ {
					img.Set(col*pixelSize+px, row*pixelSize+py, pixelColor)
				}
			}
		}
	}

	return img
}

func DrawSprite(screen *ebiten.Image, sprite *SpriteData, x, y float64) {
	if sprite == nil {
		return
	}

	// Create pre-rendered image if not exists
	if sprite.Image == nil {
		sprite.Image = createSpriteImage(sprite)
	}

	// Draw pre-rendered image (much faster!)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(sprite.Image, op)
}

func GetPlayerSprite(class PlayerClass) *SpriteData {
	return playerSprites[class]
}

func GetEnemySprite(enemyType EnemyType, level EnemyLevel) *SpriteData {
	if typeMap, exists := enemySprites[enemyType]; exists {
		if sprite, exists := typeMap[level]; exists {
			return sprite
		}
	}
	return enemySprites[Scout][Level1]
}

func GetEnemyStats(enemyType EnemyType, level EnemyLevel) EnemyStats {
	if typeMap, exists := enemyStats[enemyType]; exists {
		if stats, exists := typeMap[level]; exists {
			return stats
		}
	}
	return enemyStats[Scout][Level1]
}
