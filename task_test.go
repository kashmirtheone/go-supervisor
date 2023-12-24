package supervisor

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	t.Run("creates a new basic task", func(t *testing.T) {
		sstart := func() error { return nil }
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

		require.NotZero(t, r, "not zero value")
		require.Equal(t, "name", r.name, "right name")
	})
}

func TestTask_Start(t *testing.T) {
	t.Run("panics while starting", func(t *testing.T) {
		sstart := func() error {
			panic("puff")
			return nil
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

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
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

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
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
	})
	t.Run("returns error while starting but with retry policy", func(t *testing.T) {
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
			"eventually right state stopped",
		)
	})
	t.Run("returns error while starting but with ignore policy", func(t *testing.T) {
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
	})
	t.Run("starts with non locking callback", func(t *testing.T) {
		sstart := func() error {
			return nil
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

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
	t.Run("tries to start a task that is already started", func(t *testing.T) {
		control := make(chan struct{})
		sstart := func() error {
			<-control
			return nil
		}
		sstop := func() error {
			control <- struct{}{}
			return nil
		}
		r := NewTask("name", sstart, sstop, Options{})

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
	})
	t.Run("started with success", func(t *testing.T) {
		control := make(chan struct{})
		sstart := func() error {
			<-control
			return nil
		}
		sstop := func() error {
			control <- struct{}{}
			return nil
		}
		r := NewTask("name", sstart, sstop, Options{})

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

func TestTask_Stop(t *testing.T) {
	t.Run("stops a process that is not started yet", func(t *testing.T) {
		sstart := func() error { return nil }
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

		r.Stop()
		require.Equal(t, processStateReady, r.state, "right state")
	})
	t.Run("stops an aborted process", func(t *testing.T) {
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error {
			return nil
		}
		r := NewTask("name", sstart, sstop, Options{})

		go r.Start()
		time.Sleep(time.Millisecond)
		r.Stop()
		require.Equal(t, processStateAborted, r.state, "right state")
	})
	t.Run("stops an aborted process but with shutdown policy", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
			"eventually right state stopped",
		)
	})
	t.Run("stops an aborted process but with retry policy", func(t *testing.T) {
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error {
			return nil
		}
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
		sstart := func() error {
			return fmt.Errorf("some error")
		}
		sstop := func() error {
			return nil
		}
		r := NewTask("name", sstart, sstop, Options{Policy: PolicyOptions{
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
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
	})
	t.Run("stops with non locking callback", func(t *testing.T) {
		sstart := func() error { return nil }
		sstop := func() error { return nil }
		r := NewTask("name", sstart, sstop, Options{})

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
	t.Run("stop with error", func(t *testing.T) {
		control := make(chan struct{})
		sstart := func() error {
			<-control
			return nil
		}
		sstop := func() error {
			control <- struct{}{}
			return fmt.Errorf("some error")
		}
		r := NewTask("name", sstart, sstop, Options{})

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
				return atomic.LoadInt32(&r.state) == processStateAborted
			},
			time.Second,
			time.Millisecond,
			"eventually right state aborted",
		)
	})
}

func TestTask_Name(t *testing.T) {
	sstart := func() error { return nil }
	sstop := func() error { return nil }
	r := NewTask("name", sstart, sstop, Options{})

	require.Equal(t, "name", r.Name(), "equal name")
}
