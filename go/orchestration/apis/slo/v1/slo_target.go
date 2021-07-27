// This code was taken from the Polaris SLO Cloud project (https://polaris-slo-cloud.github.io/),
// which is licensed under the Apache 2.0 License.

package v1

// +kubebuilder:validation:Required

import (
	autoscaling "k8s.io/api/autoscaling/v1"
)

// SloTarget specifies the target entity for an elasticity strategy and SLO.
type SloTarget struct {
	// Specifies the target on which to execute the elasticity strategy.
	autoscaling.CrossVersionObjectReference `json:",inline"`
}
