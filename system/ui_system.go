package system

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/ajkula/shmup/config"
	"github.com/ajkula/shmup/state"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Star struct {
	X, Y  float32
	Speed float32
	Size  float32
}

type Particle struct {
	X, Y   float32
	VX, VY float32
	Life   float32
	MaxLife float32
	Color  color.RGBA
	Size   float32
}

type UISystem struct {
	screenWidth  int
	screenHeight int
	stars        []Star
	particles    []Particle
	menuTimer    float64
	victoryTimer float64
}

func NewUISystem() *UISystem {
	ui := &UISystem{
		screenWidth:  config.Config.ScreenWidth,
		screenHeight: config.Config.ScreenHeight,
		stars:        make([]Star, 100),
		particles:    make([]Particle, 0),
		menuTimer:    0,
		victoryTimer: 0,
	}

	// Initialize stars
	for i := range ui.stars {
		ui.stars[i] = Star{
			X:     rand.Float32() * float32(ui.screenWidth),
			Y:     rand.Float32() * float32(ui.screenHeight),
			Speed: rand.Float32()*2 + 0.5,
			Size:  rand.Float32()*2 + 1,
		}
	}

	return ui
}

func (ui *UISystem) Update(deltaTime float64, gameState state.GameState) {
	ui.menuTimer += deltaTime
	ui.victoryTimer += deltaTime

	// Update stars
	for i := range ui.stars {
		ui.stars[i].Y += ui.stars[i].Speed
		if ui.stars[i].Y > float32(ui.screenHeight) {
			ui.stars[i].Y = 0
			ui.stars[i].X = rand.Float32() * float32(ui.screenWidth)
		}
	}

	// Update particles
	if gameState == state.StateVictory {
		// Add new particles
		if rand.Float32() < 0.3 {
			for i := 0; i < 3; i++ {
				ui.addVictoryParticle()
			}
		}
	}

	// Update existing particles
	for i := len(ui.particles) - 1; i >= 0; i-- {
		p := &ui.particles[i]
		p.X += p.VX * float32(deltaTime) * 60
		p.Y += p.VY * float32(deltaTime) * 60
		p.VY += 0.2 // gravity
		p.Life -= float32(deltaTime)

		if p.Life <= 0 {
			ui.particles = append(ui.particles[:i], ui.particles[i+1:]...)
		}
	}
}

func (ui *UISystem) addVictoryParticle() {
	x := rand.Float32() * float32(ui.screenWidth)
	y := float32(ui.screenHeight) * 0.3

	angle := rand.Float32() * math.Pi * 2
	speed := rand.Float32()*3 + 2

	colors := []color.RGBA{
		{255, 200, 0, 255},   // Gold
		{255, 100, 0, 255},   // Orange
		{255, 50, 50, 255},   // Red
		{100, 200, 255, 255}, // Blue
		{200, 100, 255, 255}, // Purple
	}

	ui.particles = append(ui.particles, Particle{
		X:       x,
		Y:       y,
		VX:      float32(math.Cos(float64(angle))) * speed,
		VY:      float32(math.Sin(float64(angle)))*speed - 5,
		Life:    2.0,
		MaxLife: 2.0,
		Color:   colors[rand.Intn(len(colors))],
		Size:    rand.Float32()*3 + 2,
	})
}

func (ui *UISystem) Draw(screen *ebiten.Image, gameState state.GameState, lives, score, currentLevel, highScore, playerHealth, playerMaxHealth int, bossActive bool, bossHealth, bossMaxHealth int) {
	// Draw HI-SCORE at the top of all screens
	ui.drawHighScore(screen, highScore)

	switch gameState {
	case state.StateMainMenu:
		ui.drawMenu(screen)
	case state.StatePlaying:
		ui.drawHUD(screen, lives, score, currentLevel, playerHealth, playerMaxHealth, bossActive, bossHealth, bossMaxHealth)
	case state.StateGameOver:
		ui.drawGameOver(screen, score)
	case state.StateVictory:
		ui.drawVictory(screen, score)
	}
}

func (ui *UISystem) drawHighScore(screen *ebiten.Image, highScore int) {
	hiScoreText := fmt.Sprintf("HI-SCORE: %d", highScore)
	hiScoreX := ui.screenWidth/2 - len(hiScoreText)*6/2

	// Draw background
	vector.DrawFilledRect(screen,
		float32(hiScoreX-5), 2,
		float32(len(hiScoreText)*6+10), 16,
		color.RGBA{0, 0, 0, 200}, false)

	// Draw border
	vector.StrokeRect(screen,
		float32(hiScoreX-5), 2,
		float32(len(hiScoreText)*6+10), 16,
		1, color.RGBA{255, 215, 0, 255}, false) // Gold border

	// Draw text
	ebitenutil.DebugPrintAt(screen, hiScoreText, hiScoreX, 5)
}

func (ui *UISystem) drawStars(screen *ebiten.Image) {
	for _, star := range ui.stars {
		brightness := uint8(150 + star.Size*50)
		vector.DrawFilledCircle(screen, star.X, star.Y, star.Size, color.RGBA{brightness, brightness, brightness, 255}, false)
	}
}

func (ui *UISystem) drawMenu(screen *ebiten.Image) {
	// Draw stars background
	ui.drawStars(screen)

	// Large SHMUP title with wave color effect
	titleText := "SHMUP"
	letterWidth := float32(60)  // Large letters
	letterHeight := float32(80)
	letterSpacing := float32(10)
	totalWidth := float32(len(titleText))*letterWidth + float32(len(titleText)-1)*letterSpacing
	startX := float32(ui.screenWidth)/2 - totalWidth/2
	startY := float32(ui.screenHeight)/2 - letterHeight/2 - 60

	// Draw each letter with wave color
	for i, char := range titleText {
		letterX := startX + float32(i)*(letterWidth+letterSpacing)

		// Calculate wave color (blue -> orange -> blue)
		// Wave moves from left to right
		wavePos := (ui.menuTimer*2 + float64(i)*0.3) // Wave speed and spacing
		colorPhase := math.Sin(wavePos)

		// Interpolate between blue and orange
		var r, g, b uint8
		if colorPhase > 0 {
			// Blue to Orange
			t := colorPhase
			r = uint8(30 + t*225)  // 30 -> 255
			g = uint8(144 + t*21)  // 144 -> 165
			b = uint8(255 - t*255) // 255 -> 0
		} else {
			// Orange to Blue
			t := -colorPhase
			r = uint8(255 - t*225) // 255 -> 30
			g = uint8(165 - t*21)  // 165 -> 144
			b = uint8(0 + t*255)   // 0 -> 255
		}

		letterColor := color.RGBA{r, g, b, 255}

		// Draw glow/shadow
		for offsetX := float32(-3); offsetX <= 3; offsetX++ {
			for offsetY := float32(-3); offsetY <= 3; offsetY++ {
				if offsetX == 0 && offsetY == 0 {
					continue
				}
				dist := math.Sqrt(float64(offsetX*offsetX + offsetY*offsetY))
				glowAlpha := uint8(80 * (1.0 - dist/5.0))
				glowColor := color.RGBA{r / 2, g / 2, b / 2, glowAlpha}
				ui.drawLargeLetter(screen, char, letterX+offsetX*2, startY+offsetY*2, letterWidth, letterHeight, glowColor)
			}
		}

		// Draw main letter
		ui.drawLargeLetter(screen, char, letterX, startY, letterWidth, letterHeight, letterColor)
	}

	// Subtitle - properly centered
	subtitle := "RETRO SPACE SHOOTER"
	subtitleY := startY + letterHeight + 30
	subtitleX := ui.screenWidth/2 - len(subtitle)*6/2
	ebitenutil.DebugPrintAt(screen, subtitle, subtitleX, int(subtitleY))

	// Blinking "Press any key" - properly centered
	if int(ui.menuTimer*2)%2 == 0 {
		instructions := "PRESS ANY KEY TO START"
		instructY := subtitleY + 50
		instructX := ui.screenWidth/2 - len(instructions)*6/2
		instructColor := color.RGBA{0, 255, 0, 255}

		// Draw background for visibility
		vector.DrawFilledRect(screen,
			float32(instructX-2), float32(instructY-2),
			float32(len(instructions)*6+4), 15,
			color.RGBA{0, 0, 0, 150}, false)

		ebitenutil.DebugPrintAt(screen, instructions, instructX, int(instructY))

		// Add colored border
		vector.StrokeRect(screen,
			float32(instructX-2), float32(instructY-2),
			float32(len(instructions)*6+4), 15,
			2, instructColor, false)
	}

	// Credits - at bottom center
	credits := "MADE WITH EBITEN"
	creditsX := ui.screenWidth/2 - len(credits)*6/2
	ebitenutil.DebugPrintAt(screen, credits, creditsX, ui.screenHeight-30)
}

// Draw a large letter using rectangles
func (ui *UISystem) drawLargeLetter(screen *ebiten.Image, char rune, x, y, width, height float32, col color.Color) {
	// Simple block letters - draw filled rectangles based on the character
	switch char {
	case 'A':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)

	case 'C':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)

	case 'E':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)

	case 'G':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)
		// Middle horizontal (right half)
		vector.DrawFilledRect(screen, x+width/2, y+height*2/5, width/2, height/5, col, false)
		// Right vertical (bottom half)
		vector.DrawFilledRect(screen, x+width*4/5, y+height*2/5, width/5, height*3/5, col, false)

	case 'H':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)

	case 'I':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Center vertical
		vector.DrawFilledRect(screen, x+width*2/5, y, width/5, height, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)

	case 'M':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height, col, false)
		// Left diagonal part
		vector.DrawFilledRect(screen, x+width/5, y, width/5, height/2, col, false)
		// Right diagonal part
		vector.DrawFilledRect(screen, x+width*3/5, y, width/5, height/2, col, false)

	case 'O':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)

	case 'P':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)
		// Right vertical (top half)
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height*2/5, col, false)

	case 'R':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)
		// Right vertical (top half)
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height*2/5, col, false)
		// Diagonal (bottom right)
		vector.DrawFilledRect(screen, x+width*4/5, y+height*3/5, width/5, height*2/5, col, false)

	case 'S':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Middle horizontal
		vector.DrawFilledRect(screen, x, y+height*2/5, width, height/5, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)
		// Top-left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height*2/5, col, false)
		// Bottom-right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y+height*3/5, width/5, height*2/5, col, false)

	case 'T':
		// Top horizontal
		vector.DrawFilledRect(screen, x, y, width, height/5, col, false)
		// Center vertical
		vector.DrawFilledRect(screen, x+width*2/5, y, width/5, height, col, false)

	case 'U':
		// Left vertical
		vector.DrawFilledRect(screen, x, y, width/5, height, col, false)
		// Right vertical
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height, col, false)
		// Bottom horizontal
		vector.DrawFilledRect(screen, x, y+height*4/5, width, height/5, col, false)

	case 'V':
		// Left vertical (slightly angled)
		vector.DrawFilledRect(screen, x, y, width/5, height*4/5, col, false)
		// Right vertical (slightly angled)
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height*4/5, col, false)
		// Bottom center
		vector.DrawFilledRect(screen, x+width*2/5, y+height*4/5, width/5, height/5, col, false)

	case 'Y':
		// Top-left diagonal
		vector.DrawFilledRect(screen, x, y, width/5, height*2/5, col, false)
		// Top-right diagonal
		vector.DrawFilledRect(screen, x+width*4/5, y, width/5, height*2/5, col, false)
		// Center vertical
		vector.DrawFilledRect(screen, x+width*2/5, y+height*2/5, width/5, height*3/5, col, false)
	}
}

// Draw a large word using large letters
func (ui *UISystem) drawLargeWord(screen *ebiten.Image, word string, x, y, letterWidth, letterHeight, spacing float32, col color.Color, withGlow bool, glowColor color.RGBA) {
	currentX := x
	for _, char := range word {
		if char == ' ' {
			currentX += letterWidth / 2 // Half width for spaces
			continue
		}

		// Draw glow if requested
		if withGlow {
			for offsetX := float32(-2); offsetX <= 2; offsetX++ {
				for offsetY := float32(-2); offsetY <= 2; offsetY++ {
					if offsetX == 0 && offsetY == 0 {
						continue
					}
					ui.drawLargeLetter(screen, char, currentX+offsetX*2, y+offsetY*2, letterWidth, letterHeight, glowColor)
				}
			}
		}

		// Draw main letter
		ui.drawLargeLetter(screen, char, currentX, y, letterWidth, letterHeight, col)
		currentX += letterWidth + spacing
	}
}

func (ui *UISystem) drawHUD(screen *ebiten.Image, lives, score, currentLevel, playerHealth, playerMaxHealth int, bossActive bool, bossHealth, bossMaxHealth int) {
	// HUD positioned below HI-SCORE (Y=30 instead of Y=5)
	hudY := float32(30)
	hudTextY := int(hudY + 5)

	// Lives (top-left) with heart symbol
	livesText := fmt.Sprintf("LIVES: %d", lives)
	livesColor := color.RGBA{255, 50, 50, 255}
	if lives <= 1 {
		// Blink when low on lives
		if int(ui.menuTimer*4)%2 == 0 {
			livesColor = color.RGBA{255, 0, 0, 255}
		}
	}
	vector.DrawFilledRect(screen, 5, hudY, float32(len(livesText)*6+10), 18, color.RGBA{0, 0, 0, 150}, false)
	vector.DrawFilledRect(screen, 7, hudY+2, float32(len(livesText)*6+6), 14, livesColor, false)
	ebitenutil.DebugPrintAt(screen, livesText, 10, hudTextY)

	// Player health bar (below lives)
	playerHealthBarY := hudY + 27 // Moved down 5px
	ui.drawHealthBar(screen, 10, playerHealthBarY, 100, 6, playerHealth, playerMaxHealth, false)

	// Score (top-right)
	scoreText := fmt.Sprintf("SCORE: %d", score)
	scoreX := ui.screenWidth - len(scoreText)*6 - 20
	vector.DrawFilledRect(screen, float32(scoreX-5), hudY, float32(len(scoreText)*6+10), 18, color.RGBA{0, 0, 0, 150}, false)
	vector.DrawFilledRect(screen, float32(scoreX-3), hudY+2, float32(len(scoreText)*6+6), 14, color.RGBA{255, 200, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, hudTextY)

	// Level (top-center)
	levelText := fmt.Sprintf("LEVEL %d", currentLevel)
	levelX := ui.screenWidth/2 - len(levelText)*6/2
	vector.DrawFilledRect(screen, float32(levelX-5), hudY, float32(len(levelText)*6+10), 18, color.RGBA{0, 0, 0, 150}, false)
	vector.DrawFilledRect(screen, float32(levelX-3), hudY+2, float32(len(levelText)*6+6), 14, color.RGBA{0, 200, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, levelText, levelX, hudTextY)

	// Boss health bar (if boss is active, aligned with player health bar)
	if bossActive {
		bossHealthBarX := float32(ui.screenWidth - 210)
		ui.drawHealthBar(screen, bossHealthBarX, playerHealthBarY, 200, 6, bossHealth, bossMaxHealth, true)
	}
}

func (ui *UISystem) drawHealthBar(screen *ebiten.Image, x, y, width, height float32, currentHealth, maxHealth int, isBoss bool) {
	if maxHealth <= 0 {
		return
	}

	// Calculate fill percentage
	fillPercentage := float32(currentHealth) / float32(maxHealth)
	if fillPercentage < 0 {
		fillPercentage = 0
	}
	if fillPercentage > 1 {
		fillPercentage = 1
	}

	fillWidth := width * fillPercentage

	// Draw red background (shows lost health)
	vector.DrawFilledRect(screen, x, y, width, height, color.RGBA{255, 0, 0, 255}, false)

	// Draw yellow health fill
	if fillWidth > 0 {
		vector.DrawFilledRect(screen, x, y, fillWidth, height, color.RGBA{255, 255, 0, 255}, false)
	}

	// Draw blue border
	vector.StrokeRect(screen, x, y, width, height, 1, color.RGBA{0, 0, 255, 255}, false)

	// Draw label
	label := "PLAYER"
	if isBoss {
		label = "BOSS"
	}
	labelY := y - 12
	ebitenutil.DebugPrintAt(screen, label, int(x), int(labelY))

	// Draw health numbers
	healthText := fmt.Sprintf("%d/%d", currentHealth, maxHealth)
	textX := x + width + 5
	ebitenutil.DebugPrintAt(screen, healthText, int(textX), int(y-2))
}

func (ui *UISystem) drawGameOver(screen *ebiten.Image, score int) {
	// Dark overlay
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 200}, false)

	// Draw some stars
	ui.drawStars(screen)

	// Large GAME OVER title with red pulsing effect
	letterWidth := float32(40)
	letterHeight := float32(50)
	spacing := float32(5)

	// Calculate word widths
	gameWord := "GAME"
	overWord := "OVER"
	gameWidth := float32(len(gameWord))*(letterWidth+spacing) - spacing
	overWidth := float32(len(overWord))*(letterWidth+spacing) - spacing

	// Position
	gameX := float32(ui.screenWidth)/2 - gameWidth/2
	overX := float32(ui.screenWidth)/2 - overWidth/2
	titleY := float32(ui.screenHeight)/2 - letterHeight - 20

	// Pulsing red color
	pulse := float64(math.Sin(ui.menuTimer*4) * 0.3 + 0.7)
	titleColor := color.RGBA{255, uint8(50 * pulse), uint8(50 * pulse), 255}
	glowColor := color.RGBA{200, 0, 0, uint8(60 * pulse)}

	// Draw GAME
	ui.drawLargeWord(screen, gameWord, gameX, titleY, letterWidth, letterHeight, spacing, titleColor, true, glowColor)
	// Draw OVER
	ui.drawLargeWord(screen, overWord, overX, titleY+letterHeight+15, letterWidth, letterHeight, spacing, titleColor, true, glowColor)

	// Score
	scoreText := fmt.Sprintf("FINAL SCORE: %d", score)
	scoreY := titleY + letterHeight*2 + 50
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(scoreText)*6/2-3),
		float32(scoreY-2),
		float32(len(scoreText)*6+6), 17, color.RGBA{0, 0, 0, 200}, false)
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(scoreText)*6/2),
		float32(scoreY),
		float32(len(scoreText)*6), 13, color.RGBA{255, 200, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, scoreText, ui.screenWidth/2-len(scoreText)*6/2, int(scoreY))

	// Instructions
	if int(ui.menuTimer*2)%2 == 0 {
		instructions := "PRESS ANY KEY TO RESTART"
		instructY := scoreY + 40
		vector.DrawFilledRect(screen,
			float32(ui.screenWidth/2-len(instructions)*6/2-3),
			float32(instructY-2),
			float32(len(instructions)*6+6), 17, color.RGBA{0, 0, 0, 150}, false)
		ebitenutil.DebugPrintAt(screen, instructions, ui.screenWidth/2-len(instructions)*6/2, int(instructY))

		// Border
		vector.StrokeRect(screen,
			float32(ui.screenWidth/2-len(instructions)*6/2-3),
			float32(instructY-2),
			float32(len(instructions)*6+6), 17,
			1, color.RGBA{150, 150, 150, 255}, false)
	}
}

func (ui *UISystem) drawVictory(screen *ebiten.Image, score int) {
	// Dark background
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 10, 30, 255}, false)

	// Draw stars
	ui.drawStars(screen)

	// Draw particles
	for _, p := range ui.particles {
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := p.Color
		c.A = alpha
		vector.DrawFilledCircle(screen, p.X, p.Y, p.Size, c, false)
	}

	// Large VICTORY title with rainbow cycling effect
	letterWidth := float32(45)
	letterHeight := float32(55)
	spacing := float32(5)

	victoryWord := "VICTORY"
	victoryWidth := float32(len(victoryWord))*(letterWidth+spacing) - spacing
	victoryX := float32(ui.screenWidth)/2 - victoryWidth/2
	titleY := float32(ui.screenHeight)/2 - letterHeight - 30

	// Rainbow cycling - each letter gets a different phase
	for i, char := range victoryWord {
		// Calculate rainbow color based on timer and letter position
		hue := (ui.victoryTimer*0.5 + float64(i)*0.2)
		hue = hue - math.Floor(hue) // Keep in 0-1 range

		// Convert HSV to RGB (simple rainbow)
		var r, g, b uint8
		sector := int(hue * 6)
		f := hue*6 - float64(sector)

		switch sector % 6 {
		case 0:
			r, g, b = 255, uint8(f*255), 0
		case 1:
			r, g, b = uint8((1-f)*255), 255, 0
		case 2:
			r, g, b = 0, 255, uint8(f*255)
		case 3:
			r, g, b = 0, uint8((1-f)*255), 255
		case 4:
			r, g, b = uint8(f*255), 0, 255
		case 5:
			r, g, b = 255, 0, uint8((1-f)*255)
		}

		letterColor := color.RGBA{r, g, b, 255}
		glowColor := color.RGBA{r / 2, g / 2, b / 2, 100}

		// Position with slight wave effect
		wave := math.Sin(ui.victoryTimer*2 + float64(i)*0.5) * 5
		letterX := victoryX + float32(i)*(letterWidth+spacing)
		letterY := titleY + float32(wave)

		// Draw with glow
		ui.drawLargeWord(screen, string(char), letterX, letterY, letterWidth, letterHeight, spacing, letterColor, true, glowColor)
	}

	// Congratulations
	congrats := "ALL LEVELS COMPLETED!"
	congratsY := titleY + letterHeight + 40
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(congrats)*6/2-3),
		float32(congratsY-2),
		float32(len(congrats)*6+6), 17, color.RGBA{0, 0, 0, 200}, false)
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(congrats)*6/2),
		float32(congratsY),
		float32(len(congrats)*6), 13, color.RGBA{100, 255, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, congrats, ui.screenWidth/2-len(congrats)*6/2, int(congratsY))

	// Score
	scoreText := fmt.Sprintf("FINAL SCORE: %d", score)
	scoreY := congratsY + 35
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(scoreText)*6/2-3),
		float32(scoreY-2),
		float32(len(scoreText)*6+6), 17, color.RGBA{0, 0, 0, 200}, false)
	vector.DrawFilledRect(screen,
		float32(ui.screenWidth/2-len(scoreText)*6/2),
		float32(scoreY),
		float32(len(scoreText)*6), 13, color.RGBA{255, 255, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, scoreText, ui.screenWidth/2-len(scoreText)*6/2, int(scoreY))

	// Instructions
	if int(ui.victoryTimer*2)%2 == 0 {
		instructions := ">>> PRESS ANY KEY TO RESTART <<<"
		instructY := scoreY + 40
		vector.DrawFilledRect(screen,
			float32(ui.screenWidth/2-len(instructions)*6/2-3),
			float32(instructY-2),
			float32(len(instructions)*6+6), 17, color.RGBA{0, 0, 0, 150}, false)
		ebitenutil.DebugPrintAt(screen, instructions, ui.screenWidth/2-len(instructions)*6/2, int(instructY))

		// Cyan border
		vector.StrokeRect(screen,
			float32(ui.screenWidth/2-len(instructions)*6/2-3),
			float32(instructY-2),
			float32(len(instructions)*6+6), 17,
			1, color.RGBA{0, 255, 255, 255}, false)
	}
}
