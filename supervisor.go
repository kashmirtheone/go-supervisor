package supervisor

import (
	"context"
	"fmt"
	"sync"

	"gitlab.com/marcoxavier/supervisor/signal"
)

// Process is a process managed by supervisor.
type process interface {
	Run(ctx context.Context) error
	Name() string
}

// Supervisor is a process supervisor.
// It manages all process according failure policy.
type Supervisor struct {
	mux       sync.Mutex
	processes map[string]process
	shutdown  func() <-chan struct{}
	policy    Policy
	logger    Logger
}

// NewSupervisor creates a new supervisor with default policies.
func NewSupervisor(policyOptions ...PolicyOption) Supervisor {
	s := Supervisor{
		mux:       sync.Mutex{},
		processes: make(map[string]process),
		policy: Policy{
			Failure: FailurePolicy{
				Policy: shutdown,
			},
			Restart: RestartPolicy{
				Policy:      never,
				MaxAttempts: 1,
			},
		},
		logger: defaultLogger,
	}
	s.shutdown = signal.OSShutdownSignal
	s.policy.Reconfigure(policyOptions...)

	return s
}

// SetLogger allows you to set the logger.
func (s *Supervisor) SetLogger(logger Logger) {
	if logger != nil {
		s.logger = logger
	}
}

// DisableLogger allows you to disable the logger.
func (s *Supervisor) DisableLogger() {
	s.logger = dumbLogger
}

// SetShutdownSignal allows you to set the shutdown signal.
func (s *Supervisor) SetShutdownSignal(signal func() <-chan struct{}) {
	if signal != nil {
		s.shutdown = signal
	}
}

// AddRunner adds the callback to a runner.
// Runner will trigger callback once supervisor starts.
// Name is the identifier of runner, it need to be unique across all runners in this pool or will be ignored.
func (s *Supervisor) AddRunner(name string, callback Callback, policyOptions ...PolicyOption) {
	key := fmt.Sprintf("%s-%s", "runner", name)
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.processes[key]; exists {
		s.logger(Error, loggerData{"name": name}, "runner already exists")
		return
	}

	r := runner{
		Callback:      callback,
		name:          name,
		restartPolicy: s.policy.Restart,
		logger:        s.logger,
	}

	p := Policy{
		Restart: s.policy.Restart,
	}
	p.Reconfigure(policyOptions...)

	r.restartPolicy = p.Restart

	s.processes[key] = &r
}

// AddTask adds the task to pool.
// Task will be started once supervisor starts.
// Name is the identifier of task, it need to be unique across all tasks in this pool or will be ignored.
func (s *Supervisor) AddTask(name string, startStopper StartStopper, policyOptions ...PolicyOption) {
	key := fmt.Sprintf("%s-%s", "task", name)
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, exists := s.processes[key]; exists {
		s.logger(Error, loggerData{"name": name}, "task already exists")
		return
	}

	t := task{
		StartStopper: startStopper,
		name:         name,
		logger:       s.logger,
	}

	p := Policy{
		Restart: s.policy.Restart,
	}
	p.Reconfigure(policyOptions...)

	t.restartPolicy = p.Restart

	s.processes[key] = &t
}

// Start starts the supervisor.
func (s *Supervisor) Start() {
	s.StartWithContext(context.Background())
}

// StartWithContext starts the supervisor with a specific context.
func (s *Supervisor) StartWithContext(ctx context.Context) {
	s.logger(Info, nil, "starting supervisor")

	ctx, cancel := context.WithCancel(ctx)

	go func() {
		<-s.shutdown()
		s.logger(Info, nil, "supervisor ordered to shutdown")
		cancel()
	}()

	errChan := make(chan error)
	wg := sync.WaitGroup{}
	for name := range s.processes {
		wg.Add(1)

		process := s.processes[name]
		go func() {
			defer func() {
				if err := recover(); err != nil {
					s.logger(Warn, loggerData{"cause": err, "name": process.Name()}, "runner terminated due a panic")
				}

				wg.Done()
			}()
			if err := process.Run(ctx); err != nil {
				errChan <- err
			}
		}()
	}

	go func() {
		for range errChan {
			switch s.policy.Failure.Policy {
			case shutdown:
				s.logger(Warn, nil, "under configured failure policy, supervisor will shutdown")
				cancel()
			case ignore:
			}
		}
	}()

	wg.Wait()

	s.logger(Info, nil, "supervisor finished gracefully")
}
