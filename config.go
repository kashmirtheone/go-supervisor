package supervisor

import (
	"time"
)

// Options supervisor configuration.
type Options struct {
	Policy PolicyOptions
}

// Option is the abstract functional-parameter type used for supervisor configuration.
type Option func(*Options)

// Reconfigure reconfigures options.
func (o *Options) Reconfigure(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

// WithFailurePolicyIgnore allows you to configure the failure policy to ignore.
// Ignore failure policy ignores if process dies.
func WithFailurePolicyIgnore() Option {
	return func(o *Options) {
		o.Policy.FailurePolicy = Ignore
	}
}

// WithFailurePolicyShutdown allows you to configure the failure policy to shutdown.
// Shutdown failure policy terminates supervisor if process dies.
func WithFailurePolicyShutdown() Option {
	return func(o *Options) {
		o.Policy.FailurePolicy = Shutdown
	}
}

// WithFailurePolicyRetry allows you to configure the failure policy to retry.
// Always restarts process after they die until reaches `maxAttempts`.
// If attempts == -1 it will always restart.
func WithFailurePolicyRetry(maxAttempts int, attemptDelay time.Duration) Option {
	return func(o *Options) {
		o.Policy.FailurePolicy = Retry
		o.Policy.FailureAttempts = maxAttempts
		o.Policy.FailureDelay = attemptDelay
	}
}
