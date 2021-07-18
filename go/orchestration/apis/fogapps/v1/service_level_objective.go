package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceLevelObjective an SLOs that is attached to
// a ServiceGraph, a ServiceGraphNode, or a ServiceLink.
type ServiceLevelObjective struct {

	// Describes the type of SLO.
	SloType ApiVersionKind `json:"sloType"`

	// The name of this SLO instance.
	//
	// This must be unique within its containing list of SLOs (e.g., ServiceGraphNode.SLOs if this SLO is attached to a node).
	Name string `json:"name"`

	// The user modifiable parts of the SLO configuration.
	SloUserConfig `json:",inline"`
}

// SloUserConfig contains the user modifiable parts of the SLO configuration.
type SloUserConfig struct {

	// The elasticity strategy that should be triggered upon violations of the SLO.
	ElasticityStrategy ApiVersionKind `json:"elasticityStrategy"`

	// The SLO-specific configuration.
	// +kubebuilder:pruning:PreserveUnknownFields
	SloConfig runtime.RawExtension `json:"sloConfig"`

	// Configures the duration of the period after the last elasticity strategy execution,
	// during which the strategy will not be executed again (to avoid unnecessary scaling).
	//
	// +optional
	StabilizationWindow *StabilizationWindow `json:"stabilizationWindow,omitempty"`

	// ToDo: Make staticElasticityStrategyConfig available via ServiceGraph, if necessary.
	// Static configuration to be passed to the chosen elasticity strategy.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// StaticElasticityStrategyConfig *runtime.RawExtension `json:"staticElasticityStrategyConfig,omitempty"`
}

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
