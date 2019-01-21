package main

import (
	"context"

	"gitlab.com/marcoxavier/supervisor"
)

type Task struct {
}

func (m *Task) Start(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func (m *Task) Stop() error {
	return nil
}

type Runner struct {
}

func (r *Runner) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}

func main() {
	s := supervisor.NewSupervisor(supervisor.WithFailurePolicyShutdown(), supervisor.WithRestartPolicyNever())
	runner := Runner{}
	task := &Task{}

	s.AddRunner("some_runner", runner.Run)
	s.AddTask("some_task", task)

	s.Start()
}
