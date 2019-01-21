package supervisor

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

var defaultRateLimit = time.Second

// Callback is a runner callback that will be triggered when supervisor starts.
// It injects a context that will be canceled when supervisor shutdown.
// You should listen <-ctx.Done() to lock/unlock your callback.
type Callback func(ctx context.Context) error

// Runner contains the callback and manages the restart policies and its context.
type runner struct {
	Callback      Callback
	Name          string
	RestartPolicy RestartPolicy
	terminated    int32
	Logger        Logger
}

// Run runs the Runner.
// It creates a cancel context and manages callback error according restarting policy.
func (r *runner) Run(ctx context.Context) error {
	r.Logger(Info, loggerData{"name": r.Name}, "runner is starting")

	var err error
	rCtx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ctx.Done()
		atomic.SwapInt32(&r.terminated, 1)
		cancel()
	}()

	// rate limited to 1 seconds.
	ticker := time.NewTicker(defaultRateLimit)
	defer ticker.Stop()

	attempts := 1

loop:
	for ; ; <-ticker.C {
		err = r.Callback(rCtx)
		if err != nil {
			r.Logger(Error, loggerData{"name": r.Name, "cause": fmt.Sprintf("%+v", err)}, "failed to run runner")
		}

		if atomic.LoadInt32(&r.terminated) == 1 || attempts >= r.RestartPolicy.MaxAttempts {
			break loop
		}

		switch r.RestartPolicy.Policy {
		case never:
			break loop
		case onFailure:
			if err == nil {
				break loop
			}
			r.Logger(Info, loggerData{"name": r.Name, "attempts": attempts}, "runner is restarting")
			break
		case always:
			r.Logger(Info, loggerData{"name": r.Name, "attempts": attempts}, "runner is restarting")
			break
		}

		attempts++
	}

	r.Logger(Info, loggerData{"name": r.Name}, "runner terminated")

	return err
}
