package supervisor

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

// StartStopper is a task that contains a start and stop method.
type StartStopper interface {
	Start(ctx context.Context) error
	Stop() error
}

// Task contains the start stopper and manages the restart policies and its context.
type task struct {
	StartStopper
	name          string
	restartPolicy RestartPolicy
	terminated    int32
	logger        Logger
}

// Name returns task name.
func (t *task) Name() string {
	return t.name
}

// Run starts and stops the task.
// It creates a cancel context and manages task error according restarting policy.
func (t *task) Run(ctx context.Context) error {
	t.logger(Info, loggerData{"name": t.Name()}, "task is starting")

	var err error
	rCtx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ctx.Done()
		atomic.SwapInt32(&t.terminated, 1)
		cancel()
	}()

	// rate limited to 1 seconds.
	ticker := time.NewTicker(defaultRateLimit)
	defer ticker.Stop()

	attempts := 1

loop:
	for ; ; <-ticker.C {
		err = t.Start(rCtx)
		if err != nil {
			t.logger(Error, loggerData{"name": t.Name(), "cause": fmt.Sprintf("%+v", err)}, "failed to run task")
		}

		t.logger(Debug, nil, "task is stopping")
		if stoperr := t.Stop(); stoperr != nil {
			t.logger(Error, loggerData{"name": t.Name(), "cause": fmt.Sprintf("%+v", stoperr)}, "failed to stop task")
		}

		if atomic.LoadInt32(&t.terminated) == 1 || attempts >= t.restartPolicy.MaxAttempts {
			break loop
		}

		switch t.restartPolicy.Policy {
		case never:
			break loop
		case onFailure:
			if err == nil {
				break loop
			}
			t.logger(Info, loggerData{"name": t.Name(), "attempts": attempts}, "task is restarting")
			break
		case always:
			t.logger(Info, loggerData{"name": t.Name(), "attempts": attempts}, "task is restarting")
			break
		}

		attempts++
	}

	t.logger(Info, loggerData{"name": t.Name()}, "task terminated")

	return err
}
