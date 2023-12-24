package supervisor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// Runner contains the callback and manages the restart policies and its context.
type Runner struct {
	options  Options
	name     string
	callback func(ctx context.Context) error
	mctx     context.Context
	shutdown context.CancelFunc
	state    int32
	control  chan struct{}
}

// NewRunner creates a new runner.
func NewRunner(name string, callback func(ctx context.Context) error, defaultOptions Options) *Runner {
	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		options:  defaultOptions,
		name:     name,
		callback: callback,
		mctx:     ctx,
		shutdown: cancel,
		state:    processStateReady,
		control:  make(chan struct{}),
	}
}

// Name return runner name.
func (r *Runner) Name() string {
	return r.name
}

// Start runner.
func (r *Runner) Start() {
	l := slog.With(
		slog.String("name", r.name),
		slog.Any("labels", []string{"supervisor", "runner", "start"}),
	)

	policyController := NewPolicyController(r.options.Policy)
	start := func() {
		defer func() {
			if data := recover(); data != nil {
				atomic.StoreInt32(&r.state, processStateAborted)

				l.Error("process panicked",
					slog.String("cause", fmt.Sprintf("%v", data)),
					slog.String("stack", string(debug.Stack())),
				)
			}
		}()

		for {
			if errors.Is(r.mctx.Err(), context.Canceled) {
				break
			}

			if err := r.callback(r.mctx); err != nil {
				if policyController.Decide() == DecisionRetry {
					l.Error("process thrown an error, retrying",
						slog.Int("attempts", policyController.attempts),
					)

					time.Sleep(policyController.options.FailureDelay)
					continue
				}

				l.Error("process thrown an error, exiting",
					slog.String("cause", err.Error()),
				)

				atomic.StoreInt32(&r.state, processStateAborted)
				return
			}

			break
		}
	}

	if !atomic.CompareAndSwapInt32(&r.state, processStateReady, processStateStarted) {
		l.Warn("process thrown an error, exiting")
		return
	}

	start()
	r.control <- struct{}{}
	atomic.SwapInt32(&r.state, processStateStopped)
}

// Stop runner.
func (r *Runner) Stop() {
	if !atomic.CompareAndSwapInt32(&r.state, processStateStarted, processStateStopping) {
		return
	}

	r.shutdown()
	<-r.control
}
