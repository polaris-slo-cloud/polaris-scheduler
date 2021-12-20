package v1

import (
	sloCrds "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/slo/v1"
)

// LinkQosRequirements describes the QoS requirements that a NetworkLink should fulfill.
type LinkQosRequirements struct {
	// The type of advertised NetworkLink that is required by this ServiceLink.
	//
	// +optional
	LinkType *LinkType `json:"linkType,omitempty"`

	// The throughput requirements for the network link.
	//
	// +optional
	Throughput *NetworkThroughputRequirements `json:"throughput,omitempty"`

	// The latency requirements for the the network link.
	//
	// +optional
	Latency *NetworkLatencyRequirements `json:"latency,omitempty"`

	// The average packet loss requirements for this network link.
	//
	// +optional
	PacketLoss *NetworkPacketLossRequirements `json:"packetLoss,omitempty"`

	// Configures the elasticity strategy that should be executed when the LinkQosRequirements
	// are violated at runtime.
	// If no elasticity strategy is configured, the LinkQosRequirements are only enforced at deployment time,
	// but not at runtime.
	//
	// +optional
	ElasticityStrategy *NetworkElasticityStrategyConfig `json:"elasticityStrategy,omitempty"`
}

// NetworkThroughputRequirements describes the requirements for the speed of the NetworkLink.
type NetworkThroughputRequirements struct {
	// The minimum bandwidth of the network link in kilobits per second.
	//
	// +kubebuilder:validation:Minimum=0
	MinBandwidthKbps int64 `json:"minBandwidthKbps"`

	// The maximum variance of the bandwidth of the network link.
	//
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxBandwidthVariance *int64 `json:"maxBandwidthVariance,omitempty"`
}

// NetworkLatencyRequirements describes the requirements for the latency of a NetworkLink.
type NetworkLatencyRequirements struct {
	// The maximum end-to-end network delay (i.e., latency) of a packet sent between the two nodes, connected by this NetworkLink.
	//
	// +kubebuilder:validation:Minimum=0
	MaxPacketDelayMsec int32 `json:"maxPacketDelayMsec"`

	// The maximum variance of PacketDelayMsec (i.e., jitter).
	//
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxPacketDelayVariance *int32 `json:"maxPacketDelayVariance,omitempty"`
}

// NetworkPacketLossRequirements describes the requirements for the packet loss of a NetworkLink.
type NetworkPacketLossRequirements struct {
	// The maximum packet loss in basis points (bp).
	// 1 bp = 0.01%
	//
	// The reason for not using percent for this is that the Kubernetes API does not support
	// floating point numbers and people may need more precise packet loss information than whole percents.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	MaxPacketLossBp int32 `json:"maxPacketLossBp"`
}

// NetworkElasticityStrategyConfig configures the elasticity strategy that should be executed when the LinkQosRequirements
// are violated at runtime.
type NetworkElasticityStrategyConfig struct {

	// The elasticity strategy that should be triggered when the LinkQosRequirements are violated.
	ElasticityStrategy sloCrds.ElasticityStrategyKind `json:"elasticityStrategy"`

	// Configures the duration of the period after the last elasticity strategy execution,
	// during which the strategy will not be executed again (to avoid unnecessary scaling).
	//
	// +optional
	StabilizationWindow *sloCrds.StabilizationWindow `json:"stabilizationWindow,omitempty"`

	// ToDo: Make staticElasticityStrategyConfig available via ServiceGraph, if necessary.
	// Static configuration to be passed to the chosen elasticity strategy.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	// StaticElasticityStrategyConfig *runtime.RawExtension `json:"staticElasticityStrategyConfig,omitempty"`
}
