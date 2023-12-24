package supervisor

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// Task contains the start and stop callback and manages the restart policies.
type Task struct {
	options       Options
	name          string
	startCallback func() error
	stopCallback  func() error
	state         int32
	control       chan struct{}
}

// NewTask creates a new task.
func NewTask(name string, start, stop func() error, defaultOptions Options) *Task {
	t := &Task{
		options:       defaultOptions,
		name:          name,
		startCallback: start,
		stopCallback:  stop,
		state:         processStateReady,
		control:       make(chan struct{}),
	}

	return t
}

// Name return task name.
func (t *Task) Name() string {
	return t.name
}

// Start task.
func (t *Task) Start() {
	l := slog.With(
		slog.String("name", t.name),
		slog.Any("labels", []string{"supervisor", "task", "start"}),
	)

	policyController := NewPolicyController(t.options.Policy)
	start := func() {
		defer func() {
			if data := recover(); data != nil {
				atomic.StoreInt32(&t.state, processStateAborted)

				l.Error("process panicked",
					slog.String("cause", fmt.Sprintf("%v", data)),
					slog.String("stack", string(debug.Stack())),
				)
			}
		}()

		for {
			if atomic.LoadInt32(&t.state) == processStateStopping {
				break
			}

			if err := t.startCallback(); err != nil {
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
				atomic.StoreInt32(&t.state, processStateAborted)
				return
			}

			break
		}
	}

	if !atomic.CompareAndSwapInt32(&t.state, processStateReady, processStateStarted) {
		l.Warn("process already running")
		return
	}

	start()
	t.control <- struct{}{}
	atomic.SwapInt32(&t.state, processStateStopped)
}

// Stop task.
func (t *Task) Stop() {
	l := slog.With(
		slog.String("name", t.name),
		slog.Any("labels", []string{"supervisor", "task", "stop"}),
	)

	if !atomic.CompareAndSwapInt32(&t.state, processStateStarted, processStateStopping) {
		return
	}

	if err := t.stopCallback(); err != nil {
		l.Error("process terminated with error",
			slog.String("cause", err.Error()),
		)

		atomic.StoreInt32(&t.state, processStateAborted)
		return
	}

	<-t.control
}
