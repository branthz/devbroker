package service

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/branthz/utarrow/lib/log"
)

// Repeat performs an action asynchronously on a predetermined interval.
func Repeat(ctx context.Context, interval time.Duration, action func()) context.CancelFunc {

	// Create cancellation context first
	ctx, cancel := context.WithCancel(ctx)
	safeAction := func() {
		defer handlePanic()
		action()
	}

	// Perform the action for the first time, syncrhonously
	safeAction()
	timer := time.NewTicker(interval)
	go func() {

		for {
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
				safeAction()
			}
		}
	}()

	return cancel
}

// handlePanic handles the panic and logs it out.
func handlePanic() {
	if r := recover(); r != nil {
		log.Fatalln("async", fmt.Sprintf("panic recovered: %ss \n %s", r, debug.Stack()))
	}
}
