// This code was taken from the Polaris SLO Cloud project (https://polaris-slo-cloud.github.io/),
// which is licensed under the Apache 2.0 License.

package v1

// StabilizationWindow allows configuring the period of time that an elasticity strategy controller will
// wait after applying the strategy once, before applying it again (if the SLO is still violated), to
// avoid unnecessary scaling.
//
// For example, suppose that ScaleUpSeconds = 180 and a horizontal elasticity strategy scales out at time 't' due to an SLO violation.
// At time 't + 20 seconds' the SLO's evaluation still results in a violation, but the elasticity strategy does not scale again, because
// the stabilization window for scaling up/out has not yet passed. If the SLO evaluation at 't + 200 seconds' still results in a violation,
// the controller will scale again.
type StabilizationWindow struct {

	// The number of seconds after the previous scaling operation to wait before
	// an elasticity action that increases resources (e.g., scale up/out) or an equivalent configuration change
	// can be issued due to an SLO violation.
	//
	// +optional
	// +kubebuilder:default=60
	// +kubebuilder:validation:Minimum=0
	ScaleUpSeconds *int32 `json:"scaleUpSeconds,omitempty"`

	// The number of seconds after the previous scaling operation to wait before
	// an elasticity action that decreases resources (e.g., scale down/in) or an equivalent configuration change
	// can be issued due to an SLO violation.
	//
	// +optional
	// +kubebuilder:default=300
	// +kubebuilder:validation:Minimum=0
	ScaleDownSeconds *int32 `json:"scaleDownSeconds,omitempty"`
}
