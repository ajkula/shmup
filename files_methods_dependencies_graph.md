```javascript
main.go
└── main()
    └── game.NewGame() [game/game.go]
        ├── entities.NewPlayer() [entities/player.go]
        └── game.generateNewFormations() [game/game.go]
            └── entities.NewEnemyFormation() [entities/formation.go]
                └── entities.NewEnemy() [entities/enemy.go]

    └── ebiten.RunGame(game) [game/game.go]
        ├── game.Update() [game/game.go]
        │   ├── player.Update() [entities/player.go]
        │   ├── game.spawnNewFormation() [game/game.go]
        │   │   └── game.generateNewFormations() [game/game.go]
        │   └── formation.Update() [entities/formation.go]
        │       ├── formation.updateColumnPattern() [entities/formation.go]
        │       ├── formation.updateRowPattern() [entities/formation.go]
        │       └── formation.updateLoopPattern() [entities/formation.go]
        │
        ├── game.Draw() [game/game.go]
        │   ├── player.Draw() [entities/player.go]
        │   │   └── graphics.DrawTriangle() [graphics/draw.go]
        │   └── enemy.Draw() [entities/enemy.go]
        │       ├── graphics.DrawTriangle() [graphics/draw.go]
        │       ├── graphics.DrawCube3D() [graphics/draw.go]
        │       └── graphics.DrawCircle() [graphics/draw.go]
        │
        └── game.Layout() [game/game.go]

entities/enemy.go
├── NewEnemy()
├── Update()
└── Draw()
    ├── graphics.DrawTriangle() [graphics/draw.go]
    ├── graphics.DrawCube3D() [graphics/draw.go]
    └── graphics.DrawCircle() [graphics/draw.go]

entities/formation.go
├── NewEnemyFormation()
│   └── NewEnemy() [entities/enemy.go]
├── Update()
├── updateColumnPattern()
├── updateRowPattern()
└── updateLoopPattern()

entities/player.go
├── NewPlayer()
├── Update()
└── Draw()
    └── graphics.DrawTriangle() [graphics/draw.go]

graphics/draw.go
├── DrawTriangle()
├── DrawCube3D()
└── DrawCircle()
```