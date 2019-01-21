package supervisor

import (
	"context"
	"testing"

	"github.com/pkg/errors"

	"gitlab.com/marcoxavier/supervisor/signal"

	. "github.com/onsi/gomega"
	mock "github.com/stretchr/testify/mock"
)

func TestSupervisor_NewSupervisor_WithDefaultPolicies(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor()

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(shutdown))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(1))
	Expect(s.policy.Restart.Policy).To(Equal(never))
}

func TestSupervisor_NewSupervisor_WithRestartPolicyAlways(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor(WithRestartPolicyAlways(2))

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(shutdown))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(2))
	Expect(s.policy.Restart.Policy).To(Equal(always))
}

func TestSupervisor_NewSupervisor_WithRestartPolicyOnFailure(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor(WithRestartPolicyOnFailure(2))

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(shutdown))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(2))
	Expect(s.policy.Restart.Policy).To(Equal(onFailure))
}

func TestSupervisor_NewSupervisor_WithRestartPolicyNever(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor(WithRestartPolicyNever())

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(shutdown))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(1))
	Expect(s.policy.Restart.Policy).To(Equal(never))
}

func TestSupervisor_NewSupervisor_WithFailurePolicyIgnore(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor(WithRestartPolicyNever(), WithFailurePolicyIgnore())

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(ignore))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(1))
	Expect(s.policy.Restart.Policy).To(Equal(never))
}

func TestSupervisor_NewSupervisor_WithFailurePolicyShutdown(t *testing.T) {
	RegisterTestingT(t)

	// Assign

	// Act
	s := NewSupervisor(WithRestartPolicyNever(), WithFailurePolicyShutdown())

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
	Expect(s.logger).ToNot(BeNil())
	Expect(s.policy.Failure.Policy).To(Equal(shutdown))
	Expect(s.policy.Restart.MaxAttempts).To(Equal(1))
	Expect(s.policy.Restart.Policy).To(Equal(never))
}

func TestSupervisor_NewSupervisor_SetLogger(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := Supervisor{}

	// Act
	s.SetLogger(dumbLogger)

	// Assert
	Expect(s.logger).ToNot(BeNil())
}

func TestSupervisor_NewSupervisor_SetShutdownSignal(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := Supervisor{}

	// Act
	s.SetShutdownSignal(signal.OSShutdownSignal())

	// Assert
	Expect(s.shutdown).ToNot(BeNil())
}

func TestSupervisor_NewSupervisor_DisableLogger(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := Supervisor{}

	// Act
	s.DisableLogger()

	// Assert
	Expect(s.logger).ToNot(BeNil())
}

func TestSupervisor_AddRunner_AlreadyExistent(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	run := func(_ context.Context) error { return nil }

	// Act
	s.AddRunner("some_runner", run)
	s.AddRunner("some_runner", run)

	// Assert
	addedRunner := s.processes["runner-some_runner"].(*runner)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedRunner).ToNot(BeNil())
	Expect(addedRunner.restartPolicy.Policy).To(Equal(always))
	Expect(addedRunner.restartPolicy.MaxAttempts).To(Equal(2))
}

func TestSupervisor_AddRunner_WithDefaultPolicies(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	run := func(_ context.Context) error { return nil }

	// Act
	s.AddRunner("some_runner", run)

	// Assert
	addedRunner := s.processes["runner-some_runner"].(*runner)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedRunner).ToNot(BeNil())
	Expect(addedRunner.restartPolicy.Policy).To(Equal(always))
	Expect(addedRunner.restartPolicy.MaxAttempts).To(Equal(2))
}

func TestSupervisor_AddRunner_WithRestartPolicyOnFailure(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	run := func(_ context.Context) error { return nil }

	// Act
	s.AddRunner("some_runner", run, WithRestartPolicyOnFailure(3))

	// Assert
	addedRunner := s.processes["runner-some_runner"].(*runner)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedRunner).ToNot(BeNil())
	Expect(addedRunner.restartPolicy.Policy).To(Equal(onFailure))
	Expect(addedRunner.restartPolicy.MaxAttempts).To(Equal(3))
}

func TestSupervisor_AddTask_AlreadyExistent(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	startStopper := &MockStartStopper{}

	// Act
	s.AddTask("some_task", startStopper)
	s.AddTask("some_task", startStopper)

	// Assert
	addedTask := s.processes["task-some_task"].(*task)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedTask).ToNot(BeNil())
	Expect(addedTask.restartPolicy.Policy).To(Equal(always))
	Expect(addedTask.restartPolicy.MaxAttempts).To(Equal(2))
}

func TestSupervisor_AddTask_WithDefaultPolicies(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	startStopper := &MockStartStopper{}

	// Act
	s.AddTask("some_task", startStopper)

	// Assert
	addedTask := s.processes["task-some_task"].(*task)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedTask).ToNot(BeNil())
	Expect(addedTask.restartPolicy.Policy).To(Equal(always))
	Expect(addedTask.restartPolicy.MaxAttempts).To(Equal(2))
}

func TestSupervisor_AddTask_WithRestartPolicyOnFailure(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithRestartPolicyAlways(2))
	startStopper := &MockStartStopper{}

	// Act
	s.AddTask("some_task", startStopper, WithRestartPolicyOnFailure(3))

	// Assert
	addedTask := s.processes["task-some_task"].(*task)

	Expect(len(s.processes)).To(Equal(1))
	Expect(addedTask).ToNot(BeNil())
	Expect(addedTask.restartPolicy.Policy).To(Equal(onFailure))
	Expect(addedTask.restartPolicy.MaxAttempts).To(Equal(3))
}

func TestSupervisor_Start_WithoutErrors(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor()
	p := &mockProcess{}
	s.processes["some_process"] = p
	p.On("Run", mock.Anything).Return(nil)

	// Act
	s.Start()

	// Assert
	Expect(p.AssertExpectations(t)).To(BeTrue())
}

func TestSupervisor_Start_WithPanic_FailurePolicyIgnore(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithFailurePolicyIgnore())
	p := &mockProcess{}
	s.processes["some_process"] = p
	p.On("Run", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		panic("panic here")
	})
	p.On("Name").Return("some_process")

	// Act
	s.Start()

	// Assert
	Expect(p.AssertExpectations(t)).To(BeTrue())
}

func TestSupervisor_Start_WithErrorsFailurePolicyIgnore(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithFailurePolicyIgnore())
	p := &mockProcess{}
	s.processes["some_process"] = p
	p.On("Run", mock.Anything).Return(errors.New("some error"))

	// Act
	s.Start()

	// Assert
	Expect(p.AssertExpectations(t)).To(BeTrue())
}

func TestSupervisor_Start_WithErrorsFailurePolicyShutdown(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	s := NewSupervisor(WithFailurePolicyShutdown())
	p1 := &mockProcess{}
	p2 := &mockProcess{}
	s.processes["some_process1"] = p1
	s.processes["some_process2"] = p2
	p1.On("Run", mock.AnythingOfType("*context.cancelCtx")).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(nil)
	p2.On("Run", mock.Anything).Return(errors.New("some error"))

	// Act
	s.Start()

	// Assert
	Expect(p1.AssertExpectations(t)).To(BeTrue())
	Expect(p2.AssertExpectations(t)).To(BeTrue())
}
