package v1

// LinkQosRequirements describes the QoS requirements that a NetworkLink should fulfill.
type LinkQosRequirements struct {
	// The type of advertised NetworkLink that is required by this ServiceLink.
	//
	// +optional
	LinkType LinkType `json:"linkType,omitempty"`

	// The throughput requirements for the network link.
	//
	// +optional
	Throughput NetworkThroughputRequirements `json:"throughput,omitempty"`

	// The latency requirements for the the network link.
	//
	// +optional
	Latency NetworkLatencyRequirements `json:"latency,omitempty"`

	// The average packet loss requirements for this network link.
	//
	// +optional
	PacketLoss NetworkPacketLossRequirements `json:"packetLoss,omitempty"`
}

// NetworkThroughputRequirements describes the requirements for the speed of the NetworkLink.
type NetworkThroughputRequirements struct {
	// The minimum bandwidth of the network link in kilobits per second.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	MinBandwidthKbps int64 `json:"minBandwidthKbps"`

	// The maximum variance of the bandwidth of the network link.
	//
	// +kubebuilder:validation:Minimum=0
	// +optional
	MaxBandwidthVariance *int64 `json:"MaxBandwidthVariance,omitempty"`
}

// NetworkLatencyRequirements describes the requirements for the latency of a NetworkLink.
type NetworkLatencyRequirements struct {
	// The maximum end-to-end network delay (i.e., latency) of a packet sent between the two nodes, connected by this NetworkLink.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	MaxPacketDelayMsec int32 `json:"maxPacketDelayMsec"`

	// The maximum variance of PacketDelayMsec (i.e., jitter).
	//
	// +kubebuilder:validation:Minimum=0
	// ++optional
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
	// +kubebuilder:default=0
	MaxPacketLossBp int32 `json:"maxPacketLossBp,omitempty"`
}
