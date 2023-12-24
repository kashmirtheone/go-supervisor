package supervisor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/kashmirtheone/go-supervisor/v2/signal"
)

type Process interface {
	Name() string
	Start()
	Stop()
}

const (
	processStateReady int32 = iota
	processStateStarted
	processStateStopping
	processStateStopped
	processStateAborted
)

const (
	stateReady int32 = iota
	stateStarting
	stateStarted
	stateStopping
	stateStopped
)

// Callback is a function that will be triggered when supervisor starts.
// It injects a context that will be canceled when supervisor shutdown.
// You should listen <-ctx.Done() to lock/unlock your callback.
type Callback func(ctx context.Context) error

// StartStopper is a task that contains a start and stop method.
// Start should block while running the processes.
// Stop should signal the Start method to return. It is not required for stop to only return after start has returned.
type StartStopper interface {
	Start() error
	Stop() error
}

// Supervisor is a process manager.
// It manages all processes according restart policy.
type Supervisor struct {
	options        Options
	mux            sync.Mutex
	shutdownSignal chan struct{}
	processes      map[string]Process
	state          int32
}

// New creates a new supervisor with optional configuration.
func New(options ...Option) *Supervisor {
	s := &Supervisor{
		options: Options{
			Policy: PolicyOptions{
				FailurePolicy:   Shutdown,
				ShutdownControl: make(chan struct{}),
			},
		},
		mux:            sync.Mutex{},
		shutdownSignal: signal.SigtermSignal(),
		processes:      map[string]Process{},
		state:          stateReady,
	}

	s.Reconfigure(options...)
	return s
}

// Reconfigure reconfigures supervisor.
func (s *Supervisor) Reconfigure(options ...Option) {
	s.options.Reconfigure(options...)
}

// AddTask adds the task to supervisor.
// Task will be started once supervisor starts.
// Name is the identifier of task, it needs to be unique across all tasks or will be ignored.
func (s *Supervisor) AddTask(name string, startStopper StartStopper, options ...Option) {
	l := slog.With(
		slog.String("name", name),
		slog.Any("labels", []string{"supervisor", "add-task"}),
	)
	key := fmt.Sprintf("%s-%s", "task", name)

	if atomic.LoadInt32(&s.state) != stateReady {
		l.Warn("can not add task, supervisor not in ready state")
		return
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.processes[key]; exists {
		l.Warn("task already exists")
		return
	}

	t := NewTask(name, startStopper.Start, startStopper.Stop, s.options)
	t.options.Reconfigure(options...)

	s.processes[key] = t
}

// AddRunner adds the callback to supervisor.
// Runner will trigger callback once supervisor starts.
// Name is the identifier of runner, it needs to be unique across all runners or will be ignored.
func (s *Supervisor) AddRunner(name string, callback Callback, options ...Option) {
	l := slog.With(
		slog.String("name", name),
		slog.Any("labels", []string{"supervisor", "add-runner"}),
	)
	key := fmt.Sprintf("%s-%s", "runner", name)

	if atomic.LoadInt32(&s.state) != stateReady {
		l.Warn("can not add runner, supervisor not in ready state")
		return
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.processes[key]; exists {
		l.Warn("runner already exists")
		return
	}

	r := NewRunner(name, callback, s.options)
	r.options.Reconfigure(options...)
	s.processes[key] = r
}

// Start starts the supervisor.
func (s *Supervisor) Start() {
	l := slog.With(
		slog.Any("labels", []string{"supervisor", "start"}),
	)

	if len(s.processes) <= 0 {
		l.Warn("no processes are registered")
		return
	}

	go s.runControlHandler(l)

	// start all
	l.Info("starting all processes")
	atomic.SwapInt32(&s.state, stateStarting)

	for _, process := range s.processes {
		l.Info("starting")
		go process.Start()
		l.Debug("started")
	}

	l.Info("all processes started")
	atomic.SwapInt32(&s.state, stateStarted)

	// wait shutdown
	select {
	case <-s.shutdownSignal:
		l.Info("received shutdown signal")
	}

	// stop all processes
	l.Info("stopping the all processes")
	atomic.SwapInt32(&s.state, stateStopping)

	wg := sync.WaitGroup{}
	for _, process := range s.processes {
		wg.Add(1)
		go func(process Process) {
			l.Info("waiting for process to terminate",
				slog.String("process", process.Name()),
			)
			process.Stop()
			l.Info("process stopped",
				slog.String("process", process.Name()),
			)
			wg.Done()
		}(process)
	}
	wg.Wait()

	l.Info("supervisor terminated")
	atomic.SwapInt32(&s.state, stateStopped)
}

// Shutdown stops supervisor.
func (s *Supervisor) Shutdown() {
	l := slog.With(
		slog.Any("labels", []string{"supervisor", "shutdown"}),
	)

	if atomic.LoadInt32(&s.state) != stateStarted {
		l.Warn("supervisor not running")
		return
	}

	s.shutdownSignal <- struct{}{}
}

func (s *Supervisor) runControlHandler(l *slog.Logger) {
	go func() {
		for range s.options.Policy.ShutdownControl {
			if atomic.LoadInt32(&s.state) == stateStarted {
				l.Warn("under configured failure policy, supervisor will shutdown")
				s.shutdownSignal <- struct{}{}
			}
		}
	}()
}
