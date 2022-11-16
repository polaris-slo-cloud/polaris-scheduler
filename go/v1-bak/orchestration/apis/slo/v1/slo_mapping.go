// This code was taken from the Polaris SLO Cloud project (https://polaris-slo-cloud.github.io/),
// which is licensed under the Apache 2.0 License.

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:validation:Required

// SloMapping can be used as a base type for specific SLO mapping types by embedding it inline.
// It includes the properties required by most SLO mappings.
type SloMapping struct {
	// Specifies the target on which to execute the elasticity strategy.
	TargetRef SloTarget `json:"targetRef"`

	// Specifies the type of ElasticityStrategy to use for this SLO mapping.
	ElasticityStrategy ElasticityStrategyKind `json:"elasticityStrategy"`

	// Configures the duration of the period after the last elasticity strategy execution,
	// during which the strategy will not be executed again (to avoid unnecessary scaling).
	//
	// +optional
	StabilizationWindow *StabilizationWindow `json:"stabilizationWindow,omitempty"`

	// Any static configuration parameters, which are unknown to the SLO, but which may be required to configure
	// the chosen kind of elasticity strategy should be specified here. They will be copied over
	// into the spec of the elasticity strategy.
	//
	// // Note that the Go deepCopy() method will only be able to create shallow copies of this field,
	// because it does not know its internal structure.
	// +optional
	StaticElasticityStrategyConfig *runtime.RawExtension `json:"staticElasticityStrategyConfig,omitempty"`
}
