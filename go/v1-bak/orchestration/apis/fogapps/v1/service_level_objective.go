package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	sloCrds "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/slo/v1"
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
	ElasticityStrategy sloCrds.ElasticityStrategyKind `json:"elasticityStrategy"`

	// The SLO-specific configuration.
	// +kubebuilder:pruning:PreserveUnknownFields
	SloConfig runtime.RawExtension `json:"sloConfig"`

	// Configures the duration of the period after the last elasticity strategy execution,
	// during which the strategy will not be executed again (to avoid unnecessary scaling).
	//
	// +optional
	StabilizationWindow *sloCrds.StabilizationWindow `json:"stabilizationWindow,omitempty"`

	// Static configuration to be passed to the chosen elasticity strategy.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	StaticElasticityStrategyConfig *runtime.RawExtension `json:"staticElasticityStrategyConfig,omitempty"`
}
