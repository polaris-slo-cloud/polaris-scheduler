package v1

import (
	core "k8s.io/api/core/v1"
)

// DNSConfig represents the DNS configuration for all pods of a ServiceGraph.
type DNSConfig struct {

	// Sets the DNS policy for all pods created from this ServiceGraph.
	//
	// The possible values are: 'ClusterFirstWithHostNet', 'ClusterFirst' (= default), 'Default', or 'None'.
	//
	// +kubebuilder:validation:Enum=ClusterFirstWithHostNet;ClusterFirst;Default;None
	// +kubebuilder:default=ClusterFirst
	// +optional
	DNSPolicy core.DNSPolicy `json:"dnsPolicy,omitempty"`

	// The DNS configuration that should be merged with the DNS policy.
	core.PodDNSConfig `json:",inline"`
}
