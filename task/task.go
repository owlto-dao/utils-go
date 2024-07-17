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
			log.CtxInfof(ctx, "Canceled task...")
			return
		case <-time.After(waitSecond):
		}
	}
}

func ScheduleDailyTask(ctx context.Context, fn func(), hour, minute, second int) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.CtxErrorf(ctx, "Recovered from panic: %v, stack: %s", r, string(debug.Stack()))
			}
		}()
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, second, 0, now.Location())
			if next.Before(now) {
				next = next.Add(24 * time.Hour)
			}
			duration := next.Sub(now)
			time.Sleep(duration)
			fn()
		}
	}()
}
