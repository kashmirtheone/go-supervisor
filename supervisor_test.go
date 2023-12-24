package supervisor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	os.Exit(m.Run())
}

func TestNewSupervisor(t *testing.T) {
	t.Run("with default options", func(t *testing.T) {
		s := New()

		require.NotZero(t, s, "not zero value")
		require.Equal(t, stateReady, s.state, "right state")
	})
}

func TestSupervisor_Reconfigure(t *testing.T) {

}

type startStopper struct {
	start func() error
	stop  func() error
}

func (s startStopper) Start() error {
	return s.start()
}
func (s startStopper) Stop() error {
	return s.stop()
}

func TestSupervisor_AddTask(t *testing.T) {
	t.Run("adds task with started supervisor", func(t *testing.T) {
		sv := New()
		ss := startStopper{}
		sv.AddTask("0", ss)
		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		for i := 0; i < 100; i++ {
			sv.AddTask(strconv.Itoa(i), ss)
		}
		require.Len(t, sv.processes, 1, "right number of runners")
		sv.Shutdown()
	})
	t.Run("adds task with same name", func(t *testing.T) {
		sv := New()
		ss := startStopper{}

		for i := 0; i < 100; i++ {
			sv.AddTask("name", ss)
		}

		require.Len(t, sv.processes, 1, "right number of runners")
	})
	t.Run("adds task with success", func(t *testing.T) {
		sv := New()
		ss := startStopper{}

		wg := sync.WaitGroup{}
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func(i int) {
				sv.AddTask(strconv.Itoa(i), ss)
				wg.Done()
			}(i)
		}
		wg.Wait()

		require.Len(t, sv.processes, 100, "right number of runners")
	})
}

func TestSupervisor_AddRunner(t *testing.T) {
	t.Run("adds runner with started supervisor", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error { return nil }
		sv.AddRunner("0", callback)
		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}
		require.Len(t, sv.processes, 1, "right number of runners")
		sv.Shutdown()
	})
	t.Run("adds runner with same name", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error { return nil }

		for i := 0; i < 100; i++ {
			sv.AddRunner("name", callback)
		}

		require.Len(t, sv.processes, 1, "right number of runners")
	})
	t.Run("adds runner with success", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error { return nil }

		wg := sync.WaitGroup{}
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func(i int) {
				sv.AddRunner(strconv.Itoa(i), callback)
				wg.Done()
			}(i)
		}
		wg.Wait()

		require.Len(t, sv.processes, 100, "right number of runners")
	})
}

func TestSupervisor_Start(t *testing.T) {
	t.Run("started with no processes", func(t *testing.T) {
		sv := New()

		sv.Start()
		require.Equal(t, stateReady, sv.state, "not started")
	})
	t.Run("starts all processes with error", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
		sv.Shutdown()
	})
	t.Run("starts all processes with error but with shutdown policy", func(t *testing.T) {
		sv := New(WithFailurePolicyShutdown())
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
		sv.Shutdown()
	})
	t.Run("starts all processes with error but with retry policy", func(t *testing.T) {
		sv := New(WithFailurePolicyRetry(-1, time.Millisecond*500))
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()

		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		sv.Shutdown()
	})
	t.Run("starts all processes with error but with ignore policy", func(t *testing.T) {
		sv := New(WithFailurePolicyIgnore())
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()

		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
		sv.Shutdown()
	})
	t.Run("starts all processes and locks", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error {
			<-ctx.Done()
			time.Sleep(time.Millisecond)
			return nil
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()

		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)
		sv.Shutdown()
	})
}

func TestSupervisor_Shutdown(t *testing.T) {
	t.Run("shutdown with unstarted supervisor", func(t *testing.T) {
		sv := New()

		sv.Shutdown()
		require.Equal(t, stateReady, sv.state, "not started")
	})
	t.Run("shutdown with started supervisor but no processes added", func(t *testing.T) {
		sv := New()

		go sv.Start()
		time.Sleep(time.Millisecond)

		sv.Shutdown()
		require.Equal(t, stateReady, sv.state, "not started")
	})
	t.Run("shutdown with processes", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}

		for i := 0; i < 10; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("shutdown with non-locked processes", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error {
			return nil
		}

		for i := 0; i < 10; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("shutdown with failed processes", func(t *testing.T) {
		sv := New()
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 10; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()

		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("shutdown with failed processes with shutdown policy", func(t *testing.T) {
		sv := New(WithFailurePolicyShutdown())
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 100; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("shutdown with failed processes with retry policy", func(t *testing.T) {
		sv := New(WithFailurePolicyRetry(-1, time.Millisecond*500))
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 10; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
	t.Run("shutdown with failed processes with ignore policy", func(t *testing.T) {
		sv := New(WithFailurePolicyIgnore())
		callback := func(ctx context.Context) error {
			return fmt.Errorf("some error")
		}

		for i := 0; i < 10; i++ {
			sv.AddRunner(strconv.Itoa(i), callback)
		}

		go sv.Start()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStarted
			},
			time.Second,
			time.Millisecond,
			"eventually right state started",
		)

		sv.Shutdown()
		require.Eventuallyf(t,
			func() bool {
				return atomic.LoadInt32(&sv.state) == stateStopped
			},
			time.Second,
			time.Millisecond,
			"eventually right state stopped",
		)
	})
}
