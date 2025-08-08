package system

import (
	"context"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type InputSystem struct {
	core.BaseSystem
	eventManager interfaces.EventManagerInterface
}

type InputState struct {
	Left  bool
	Right bool
	Up    bool
	Down  bool
	Shoot bool
	Pause bool
}

func NewInputSystem(eventManager interfaces.EventManagerInterface) *InputSystem {
	return &InputSystem{
		eventManager: eventManager,
	}
}

func (is *InputSystem) Initialize(ctx context.Context) error {
	return is.BaseSystem.Initialize(ctx)
}

func (is *InputSystem) Update(deltaTime float64) error {
	select {
	case <-is.CTX.Done():
		return is.CTX.Err()
	default:
		is.processInput(deltaTime)
		return nil
	}
}

func (is *InputSystem) processInput(deltaTime float64) {
	inputState := InputState{
		Left:  ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		Right: ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		Up:    ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW),
		Down:  ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS),
		Shoot: inpututil.IsKeyJustPressed(ebiten.KeySpace),
		Pause: inpututil.IsKeyJustPressed(ebiten.KeyP),
	}

	if inputState.Left || inputState.Right || inputState.Up || inputState.Down {
		is.eventManager.Publish(interfaces.InputEvent, map[string]any{
			"type":       "movement",
			"left":       inputState.Left,
			"right":      inputState.Right,
			"up":         inputState.Up,
			"down":       inputState.Down,
			"delta_time": deltaTime,
		})
	}

	if inputState.Shoot {
		is.eventManager.Publish(interfaces.InputEvent, map[string]any{
			"type": "shoot",
		})
	}

	if inputState.Pause {
		is.eventManager.Publish(interfaces.InputEvent, map[string]any{
			"type": "pause",
		})
	}
}

func (is *InputSystem) Shutdown() {
	is.BaseSystem.Shutdown()
}
