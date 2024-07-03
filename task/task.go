package task

import (
	"context"
	"time"

	"github.com/owlto-dao/utils-go/log"
	"runtime/debug"
)

func RunTask(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic: %v, stack: %s", r, string(debug.Stack()))
			}
		}()
		fn()
	}()
}

func PeriodicTask(ctx context.Context, task func(), waitSecond time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.CtxErrorf(ctx, "Recovered from panic: %v, stack: %s", r, string(debug.Stack()))
		}
	}()
	for {
		task()
		select {
		case <-ctx.Done():
			return
		case <-time.After(waitSecond):
		}
	}
}
