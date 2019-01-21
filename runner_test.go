package supervisor

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	. "github.com/onsi/gomega"
)

func init() {
	defaultRateLimit = 1
}

func TestRunner_Run_WithError_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return errors.New("some error")
		},
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Run_WithError_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return errors.New("some error")
		},
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(calls).To(Equal(2))
}

func TestRunner_Run_WithError_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return errors.New("some error")
		},
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).To(HaveOccurred())
	Expect(calls).To(Equal(2))
}

func TestRunner_Run_RunAndExitWithoutWait_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Run_RunAndExitWithoutWait_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Run_RunAndExitWithoutWait_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx := context.TODO()
	r := runner{
		Callback: func(ctx context.Context) error {
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(2))
}

func TestRunner_Run_RunAndWait_NeverPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx, cancel := context.WithCancel(context.TODO())
	r := runner{
		Callback: func(ctx context.Context) error {
			<-ctx.Done()
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      never,
			MaxAttempts: 100,
		},
		logger: dumbLogger,
	}

	// Act
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Run_RunAndWait_OnFailurePolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx, cancel := context.WithCancel(context.TODO())
	r := runner{
		Callback: func(ctx context.Context) error {
			<-ctx.Done()
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      onFailure,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	go func() {
		time.Sleep(4 * time.Second)
		cancel()
	}()

	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Run_RunAndWait_AlwaysPolicy(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	calls := 0
	ctx, cancel := context.WithCancel(context.TODO())
	r := runner{
		Callback: func(ctx context.Context) error {
			<-ctx.Done()
			calls++
			return nil
		},
		restartPolicy: RestartPolicy{
			Policy:      always,
			MaxAttempts: 2,
		},
		logger: dumbLogger,
	}

	// Act
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()

	err := r.Run(ctx)

	// Assert
	Expect(err).ToNot(HaveOccurred())
	Expect(calls).To(Equal(1))
}

func TestRunner_Name(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	r := runner{
		name: "some_name",
	}

	// Act
	name := r.Name()

	// Assert
	Expect(name).To(Equal(r.name))
}
