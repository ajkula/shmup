package util

import (
	"context"
)

func NewGameContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

func DoneChannel(ctx context.Context) <-chan struct{} {
	return ctx.Done()
}
