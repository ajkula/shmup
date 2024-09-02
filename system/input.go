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
	accumulator  float64
}

func NewInputSystem(eventManager interfaces.EventManagerInterface) *InputSystem {
	return &InputSystem{
		eventManager: eventManager,
		accumulator:  0,
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
		is.accumulator += deltaTime

		for is.accumulator >= fixedDeltaTime {
			is.processInput()
			is.accumulator -= fixedDeltaTime
		}
	}
	return nil
}

func (is *InputSystem) processInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		is.eventManager.Publish(interfaces.InputEvent, "shoot")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		is.eventManager.Publish(interfaces.InputEvent, "up")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		is.eventManager.Publish(interfaces.InputEvent, "down")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		is.eventManager.Publish(interfaces.InputEvent, "left")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		is.eventManager.Publish(interfaces.InputEvent, "right")
	}
}

func (is *InputSystem) Run(ctx context.Context) error {
	return is.BaseSystem.Run(ctx)
}

func (is *InputSystem) Shutdown() {
	// cleanup
}
