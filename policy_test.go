package supervisor

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestPolicy_Reconfigure_WithFailurePolicyIgnore(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	p := Policy{}

	// Act
	p.Reconfigure(WithFailurePolicyIgnore())

	// Assert
	Expect(p.Failure.Policy).To(Equal(ignore))
}

func TestPolicy_Reconfigure_WithFailurePolicyShutdown(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	p := Policy{}

	// Act
	p.Reconfigure(WithFailurePolicyShutdown())

	// Assert
	Expect(p.Failure.Policy).To(Equal(shutdown))
}

func TestPolicy_Reconfigure_WithRestartPolicyNever(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	p := Policy{}

	// Act
	p.Reconfigure(WithRestartPolicyNever())

	// Assert
	Expect(p.Restart.Policy).To(Equal(never))
}

func TestPolicy_Reconfigure_WithRestartPolicyAlways(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	p := Policy{}

	// Act
	p.Reconfigure(WithRestartPolicyAlways(2))

	// Assert
	Expect(p.Restart.Policy).To(Equal(always))
	Expect(p.Restart.MaxAttempts).To(Equal(2))
}

func TestPolicy_Reconfigure_WithRestartPolicyOnFailure(t *testing.T) {
	RegisterTestingT(t)

	// Assign
	p := Policy{}

	// Act
	p.Reconfigure(WithRestartPolicyOnFailure(2))

	// Assert
	Expect(p.Restart.Policy).To(Equal(onFailure))
	Expect(p.Restart.MaxAttempts).To(Equal(2))
}
