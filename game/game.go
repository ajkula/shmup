package game

import (
	"fmt"

	"github.com/ajkula/shmup/entities"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	ScreenWidth   = 640
	ScreenHeight  = 928
	EnemyCount    = 10
	BossThreshold = 50 // Nombre d'ennemis à détruire avant l'apparition du boss
)

type Game struct {
	player                *entities.Player
	enemyFormations       []*entities.EnemyFormation
	currentFormation      *entities.EnemyFormation
	boss                  *entities.Boss
	score                 int
	level                 int
	gameState             int
	elapsedTime           float64
	formationTimer        float64
	formationInterval     float64
	destroyedEnemiesCount int
}

func NewGame() *Game {
	g := &Game{
		player:            entities.NewPlayer(ScreenWidth/2, ScreenHeight-50),
		enemyFormations:   make([]*entities.EnemyFormation, 0),
		formationInterval: 3.0, // 3 secondes entre chaque formation
		formationTimer:    0,
	}

	g.generateNewFormations()

	return g
}

func (g *Game) Update() error {
	deltaTime := 1.0 / 60.0
	g.elapsedTime += deltaTime
	g.formationTimer += deltaTime

	g.player.Update(deltaTime, ScreenWidth, ScreenHeight)

	if g.currentFormation == nil {
		if g.formationTimer >= g.formationInterval {
			g.spawnNewFormation()
			g.formationTimer = 0
		}
	} else {
		formationComplete := g.currentFormation.Update(deltaTime)
		if formationComplete {
			g.currentFormation = nil
			g.formationTimer = 0
			g.spawnNewFormation()
		}
	}

	fmt.Printf("Current formation: %v, Formation timer: %.2f\n", g.currentFormation != nil, g.formationTimer)
	return nil
}

func (g *Game) spawnNewFormation() {
	if len(g.enemyFormations) == 0 {
		g.generateNewFormations()
	}
	if len(g.enemyFormations) > 0 {
		g.currentFormation = g.enemyFormations[0]
		g.enemyFormations = g.enemyFormations[1:]
		fmt.Println("Spawned new formation") // Log pour le débogage
	} else {
		fmt.Println("No formations available to spawn")
	}
}

func (g *Game) generateNewFormations() {
	g.enemyFormations = append(g.enemyFormations,
		entities.NewEnemyFormation(entities.ColumnPattern, 6, ScreenWidth/2, -100, 80),
		entities.NewEnemyFormation(entities.RowPattern, 6, ScreenWidth/2, -100, 100),
		entities.NewEnemyFormation(entities.LoopPattern, 6, ScreenWidth/2, -200, 40),
	)
	fmt.Println("Generated new formations")
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.player.Draw(screen)

	if g.currentFormation != nil && len(g.currentFormation.Enemies) > 0 {
		enemiesDrawn := 0
		for i, e := range g.currentFormation.Enemies {
			if e != nil && e.Y >= 0 && e.Y < ScreenHeight {
				e.Draw(screen)
				enemiesDrawn++
				fmt.Printf("Drawing enemy %d at (%.2f, %.2f)\n", i, e.X, e.Y)
			}
		}
		fmt.Printf("Drawing formation: %d enemies drawn\n", enemiesDrawn)
	}

	enemyCount := 0
	if g.currentFormation != nil {
		enemyCount = len(g.currentFormation.Enemies)
	}
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Time: %.2f, Enemies: %d", g.elapsedTime, enemyCount))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) enemyDestroyed() {
	g.destroyedEnemiesCount++
	g.score += 10 // Augmenter le score quand un ennemi est détruit

	if g.destroyedEnemiesCount >= BossThreshold {
		g.spawnBoss()
	}
}

func (g *Game) spawnBoss() {
	// Créer et initialiser le boss
	g.boss = entities.NewBoss(ScreenWidth/2, -100) // Supposons que NewBoss existe dans le package entities
	g.destroyedEnemiesCount = 0                    // Réinitialiser le compteur
}

func (g *Game) updateBoss(deltaTime float64) {
	if g.boss != nil {
		g.boss.Update(deltaTime)
		// Vérifier si le boss est vaincu
		if g.boss.IsDead() {
			g.boss = nil
			g.level++       // Augmenter le niveau après avoir vaincu le boss
			g.score += 1000 // Bonus de score pour avoir vaincu le boss
		}
	}
}
