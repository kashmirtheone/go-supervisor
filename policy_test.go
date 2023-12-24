package supervisor

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewPolicyController(t *testing.T) {
	c := NewPolicyController(PolicyOptions{})

	require.NotZero(t, c, "not zero value")
}

func TestPolicyController_Decide(t *testing.T) {
	t.Run("decision exit when ignore policy", func(t *testing.T) {
		c := NewPolicyController(PolicyOptions{
			FailurePolicy: Ignore,
		})

		decision := c.Decide()

		require.Equal(t, DecisionExit, decision, "decision exit")
	})
	t.Run("decision exit when shutdown policy", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		c := NewPolicyController(PolicyOptions{
			FailurePolicy:   Shutdown,
			ShutdownControl: control,
		})

		decision := c.Decide()

		require.Equal(t, DecisionExit, decision, "decision exit")
	})
	t.Run("decision exit when retry policy and attempts reached max", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		c := NewPolicyController(PolicyOptions{
			FailurePolicy:   Retry,
			ShutdownControl: control,
			FailureAttempts: 2,
		})

		decision := c.Decide()
		require.Equal(t, DecisionRetry, decision, "decision retry")
		decision = c.Decide()
		require.Equal(t, DecisionExit, decision, "decision exit")
	})
	t.Run("decision exit when retry policy and attempts is -1", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		c := NewPolicyController(PolicyOptions{
			FailurePolicy:   Retry,
			ShutdownControl: control,
			FailureAttempts: -1,
		})

		for i := 1; i < +100; i++ {
			decision := c.Decide()
			require.Equalf(t, DecisionRetry, decision, "decision retry [times:%s]", i)
		}
	})
	t.Run("decision exit when shutdown policy and control already called", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		c := NewPolicyController(PolicyOptions{
			FailurePolicy:   Shutdown,
			ShutdownControl: control,
		})

		for i := 1; i < +100; i++ {
			decision := c.Decide()
			require.Equalf(t, DecisionExit, decision, "decision exit [times:%s]", i)
		}
	})
	t.Run("decision exit when retry policy and reached max attempts and control already called", func(t *testing.T) {
		control := make(chan struct{})
		go func() {
			<-control
		}()
		c := NewPolicyController(PolicyOptions{
			FailurePolicy:   Retry,
			FailureAttempts: 1,
			ShutdownControl: control,
		})

		for i := 1; i < +100; i++ {
			decision := c.Decide()
			require.Equalf(t, DecisionExit, decision, "decision exit [times:%s]", i)
		}
	})
}
