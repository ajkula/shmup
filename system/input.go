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
	is.CTX = ctx
	return nil
}

func (is *InputSystem) Update(deltaTime float64) error {
	select {
	case <-is.CTX.Done():
		return is.CTX.Err()
	default:
		is.processInput(deltaTime)
	}
	return nil
}

func (is *InputSystem) processInput(deltaTime float64) {
	// long press
	inputState := InputState{
		Left:  ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA),
		Right: ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD),
		Up:    ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW),
		Down:  ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS),
		Shoot: ebiten.IsKeyPressed(ebiten.KeySpace),
		Pause: inpututil.IsKeyJustPressed(ebiten.KeyP),
	}

	if inputState.Left || inputState.Right || inputState.Up || inputState.Down {
		is.eventManager.Publish(interfaces.InputEvent, map[string]interface{}{
			"type":       "movement",
			"left":       inputState.Left,
			"right":      inputState.Right,
			"up":         inputState.Up,
			"down":       inputState.Down,
			"delta_time": deltaTime,
		})
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		is.eventManager.Publish(interfaces.InputEvent, map[string]interface{}{
			"type": "shoot",
		})
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		is.eventManager.Publish(interfaces.InputEvent, map[string]interface{}{
			"type": "pause",
		})
	}
}

func (is *InputSystem) Run(ctx context.Context) error {
	return is.BaseSystem.Run(ctx)
}

func (is *InputSystem) Shutdown() {
	// cleanup
}
