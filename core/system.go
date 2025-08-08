package core

import "context"

const FixedDeltaTime = 1.0 / 60.0

type System interface {
	Initialize(ctx context.Context) error
	Update(deltaTime float64) error
	Shutdown()
}

type BaseSystem struct {
	CTX context.Context
}

func (bs *BaseSystem) Initialize(ctx context.Context) error {
	bs.CTX = ctx
	return nil
}

func (bs *BaseSystem) Update(deltaTime float64) error {
	return nil
}

func (bs *BaseSystem) Shutdown() {}
