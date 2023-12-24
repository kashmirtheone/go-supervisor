package supervisor

import (
	"time"
)

const (
	DecisionRetry = iota
	DecisionExit
)

const (
	Shutdown = iota
	Ignore
	Retry
)

// PolicyOptions hold policy configuration.
type PolicyOptions struct {
	FailureAttempts int
	FailureDelay    time.Duration
	FailurePolicy   int
	ShutdownControl chan struct{}
}

// PolicyOption is the abstract functional-parameter type used for policy configuration.
type PolicyOption func(*Options)

type PolicyController struct {
	attempts int
	options  PolicyOptions
}

func NewPolicyController(options PolicyOptions) *PolicyController {
	return &PolicyController{
		options: options,
	}
}

func (c *PolicyController) Decide() int {
	switch c.options.FailurePolicy {
	case Ignore:
		return DecisionExit
	case Shutdown:
		select {
		case c.options.ShutdownControl <- struct{}{}:
		default:
		}
		return DecisionExit
	case Retry:
		c.attempts++
		if c.options.FailureAttempts == -1 || c.attempts < c.options.FailureAttempts {
			return DecisionRetry
		}
		select {
		case c.options.ShutdownControl <- struct{}{}:
		default:
		}
		return DecisionExit
	}

	return DecisionExit
}
