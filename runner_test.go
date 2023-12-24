package supervisor

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewRunner(t *testing.T) {
	t.Run("creates a new basic runner", func(t *testing.T) {
		callback := func(ctx context.Context) error { return nil }
		r := NewRunner("name", callback, Options{})

		require.NotZero(t, r, "not zero value")
		require.Equal(t, "name", r.name, "right name")
	})
}

func TestRunner_Start(t *testing.T) {
	t.Run("panics while starting", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			panic("puff")
			return nil
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
	})
	t.Run("returns error while starting", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
	})
	t.Run("returns error while starting but with shutdown policy", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailurePolicy:   Shutdown,
			ShutdownControl: control,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
	})
	t.Run("returns error while starting but with retry policy", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailureAttempts: -1,
			FailureDelay:    time.Millisecond * 500,
			FailurePolicy:   Retry,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
	})
	t.Run("returns error while starting but with ignore policy", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailurePolicy: Ignore,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
		r.Stop()
	})
	t.Run("starts with non locking callback", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return nil
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("tries to start a runner that is already started", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Start()
		r.Start()
		r.Start()
		r.Start()
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
	})
	t.Run("started with success", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
	})
}

func TestRunner_Stop(t *testing.T) {
	t.Run("stops a process that is not started yet", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}
		r := NewRunner("name", callback, Options{})

		r.Stop()
		require.Equal(t, processStateReady, r.state, "right state")
	})
	t.Run("stops an aborted process", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
		r.Stop()
		require.Equal(t, processStateAborted, r.state, "right state")
	})
	t.Run("stops an aborted process but with shutdown policy", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailurePolicy:   Shutdown,
			ShutdownControl: control,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
	})
	t.Run("stops an aborted process but with retry policy", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailureAttempts: -1,
			FailureDelay:    time.Millisecond * 500,
			FailurePolicy:   Retry,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
	})
	t.Run("stops an aborted process but with ignore policy", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}
		r := NewRunner("name", callback, Options{Policy: PolicyOptions{
			FailurePolicy: Ignore,
		}})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
	})
	t.Run("stops with non locking callback", func(t *testing.T) {
		callback := func(ctx context.Context) error {
			return nil
		}
		r := NewRunner("name", callback, Options{})

		go r.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		r.Stop()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
}

func TestRunner_Name(t *testing.T) {
	callback := func(ctx context.Context) error { return nil }
	r := NewRunner("name", callback, Options{})

	require.Equal(t, "name", r.Name(), "equal name")
}
