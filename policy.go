package supervisor

const (
	// Shutdown policy terminates whole pool if a process terminates with error.
	shutdown = iota
	//Ignore policy terminates only when all processes terminates.
	ignore

	// Never policy never restarts processes.
	never = iota
	// Always restarts process after they terminate.
	always
	// OnFailure restarts process only if it terminates with error.
	onFailure
)

// PolicyOption is the abstract functional-parameter type used for policy configuration.
type PolicyOption func(*Policy)

// WithFailurePolicyIgnore allows you to configure the failure policy to ignore..
// Ignore policy terminates only when all processes terminates.
func WithFailurePolicyIgnore() PolicyOption {
	return func(s *Policy) {
		s.Failure.Policy = ignore
	}
}

// WithFailurePolicyShutdown allows you to configure the failure policy to shutdown.
// Shutdown policy terminates whole pool if a process terminates with error.
func WithFailurePolicyShutdown() PolicyOption {
	return func(s *Policy) {
		s.Failure.Policy = shutdown
	}
}

// WithRestartPolicyAlways allows you to configure the restart policy to always.
// Always restarts process after they terminate
func WithRestartPolicyAlways(attempts int) PolicyOption {
	return func(r *Policy) {
		r.Restart.Policy = always
		if attempts > 1 {
			r.Restart.MaxAttempts = attempts
		}
	}
}

// WithRestartPolicyOnFailure allows you to configure the restart policy to on failure.
// OnFailure restarts process only if it terminates with error.
func WithRestartPolicyOnFailure(attempts int) PolicyOption {
	return func(r *Policy) {
		r.Restart.Policy = onFailure

		if attempts > 1 {
			r.Restart.MaxAttempts = attempts
		}
	}
}

// WithRestartPolicyNever allows you to configure the restart policy to never.
// Never policy never restarts processes.
func WithRestartPolicyNever() PolicyOption {
	return func(r *Policy) {
		r.Restart.Policy = never
	}
}

// FailurePolicy describes a failure policy.
type FailurePolicy struct {
	Policy int
}

// RestartPolicy describes a restart policy.
type RestartPolicy struct {
	Policy      int
	MaxAttempts int
}

// Policy aggregates all policies.
type Policy struct {
	Failure FailurePolicy
	Restart RestartPolicy
}

// Reconfigure reconfigures default policies.
func (p *Policy) Reconfigure(policyOptions ...PolicyOption) {
	for _, option := range policyOptions {
		option(p)
	}
}
