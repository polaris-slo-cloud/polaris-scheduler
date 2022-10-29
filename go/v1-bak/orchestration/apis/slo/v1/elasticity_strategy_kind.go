// This code was taken from the Polaris SLO Cloud project (https://polaris-slo-cloud.github.io/),
// which is licensed under the Apache 2.0 License.

package v1

// +kubebuilder:validation:Required

// ElasticityStrategyKind references a particular kind of ElasticityStrategy that should be used by an SloMapping.
type ElasticityStrategyKind struct {

	// The API group and version of the ElasticityStrategy.
	APIVersion string `json:"apiVersion"`

	// The kind of ElasticityStrategy.
	Kind string `json:"kind"`
}
