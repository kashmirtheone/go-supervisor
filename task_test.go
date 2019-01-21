package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	mock "github.com/stretchr/testify/mock"

	. "github.com/onsi/gomega"
)

func init() {
	defaultRateLimit = 1
}

func TestTask_Start_WithError_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(errors.New("some_error")).Times(1)
	startStopper.On("Stop").Return(nil).Times(1)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_WithError_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(errors.New("some_error")).Times(2)
	startStopper.On("Stop").Return(nil).Times(2)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_WithError_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(errors.New("some_error")).Times(2)
	startStopper.On("Stop").Return(nil).Times(2)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndExitWithoutWait_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(nil).Times(1)
	startStopper.On("Stop").Return(nil).Times(1)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndExitWithoutWait_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(nil).Times(1)
	startStopper.On("Stop").Return(nil).Times(1)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndExitWithoutWait_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.Anything).Return(nil).Times(2)
	startStopper.On("Stop").Return(nil).Times(2)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndWait_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx, cancel := context.WithCancel(context.TODO())
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil).Times(1).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	})
	startStopper.On("Stop").Return(nil).Times(1)

	// Act
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndWait_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx, cancel := context.WithCancel(context.TODO())
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil).Times(1).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	})
	startStopper.On("Stop").Return(nil).Times(1)

	// Act
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Start_StartAndWait_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx, cancel := context.WithCancel(context.TODO())
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}
	startStopper.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	})
	startStopper.On("Stop").Return(nil)

	// Act
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Stop_WithError_Logs(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	ctx := context.TODO()
	calls := 0
	startStopper := &MockStartStopper{}
	r := task{
		StartStopper: startStopper,
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: func(_ string, _ map[string]interface{}, msg string, _ ...interface{}) {
			if msg == "failed to stop task" {
				calls++
			}
		},
	}
	startStopper.On("Start", mock.Anything).Return(nil).Times(1)
	startStopper.On("Stop").Return(errors.New("some error")).Times(1)

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
	Expect(startStopper.AssertExpectations(t)).To(BeTrue())
}

func TestTask_Name(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	r := task{
		name: "some_name",
	}

	// Act
	name := r.Name()

	// Assert
	Expect(name).To(Equal(r.name))
}
