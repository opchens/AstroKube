package utils

import (
	"context"
	"sync"
)

type contextWaitGroup struct {
}

var cwg contextWaitGroup = contextWaitGroup{}

func ContextStart(ctx context.Context) {
	v := ctx.Value(cwg)
	if v == nil {
		return
	}
	wg := v.(*sync.WaitGroup)
	wg.Add(1)
}

func ContextStop(ctx context.Context) {
	v := ctx.Value(cwg)
	if v == nil {
		return
	}
	wg := v.(*sync.WaitGroup)
	wg.Done()
}

func ContextInit(ctx context.Context) context.Context {
	wg := &sync.WaitGroup{}
	return context.WithValue(ctx, cwg, wg)
}

func ContextShutdown(ctx context.Context) {
	v := ctx.Value(cwg)
	if v == nil {
		return
	}
	wg := v.(*sync.WaitGroup)
	wg.Wait()
}
