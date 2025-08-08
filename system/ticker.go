package system

import (
	"context"

	"github.com/ajkula/shmup/core"
	"github.com/ajkula/shmup/interfaces"
)

type SystemTicker struct {
	core.BaseSystem
	eventManager interfaces.EventManagerInterface
	tickData     struct{}
}

func NewSystemTicker(eventManager interfaces.EventManagerInterface) *SystemTicker {
	return &SystemTicker{
		eventManager: eventManager,
		tickData:     struct{}{},
	}
}

func (st *SystemTicker) Initialize(ctx context.Context) error {
	return st.BaseSystem.Initialize(ctx)
}

func (st *SystemTicker) Update(deltaTime float64) error {
	select {
	case <-st.CTX.Done():
		return st.CTX.Err()
	default:
		return st.eventManager.Publish(interfaces.SystemTick, st.tickData)
	}
}

func (st *SystemTicker) Shutdown() {
	st.BaseSystem.Shutdown()
}
