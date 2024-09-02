package core

import "context"

const FixedDeltaTime = 1.0 / 60.0

type System interface {
	Initialize(ctx context.Context) error
	Update(deltaTime float64) error
	Run(ctx context.Context) error
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

func (bs *BaseSystem) Run(ctx context.Context) error {
	bs.CTX = ctx
	for {
		select {
		case <-bs.CTX.Done():
			return bs.CTX.Err()
		default:
			if err := bs.Update(FixedDeltaTime); err != nil {
				return err
			}
		}
	}
}

func (bs *BaseSystem) Shutdown() {}
