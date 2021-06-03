package v1

// All fields in these CRDs are required, except if marked as `// +optional`
// +kubebuilder:validation:Required

// NetworkLinkQos describes the quality of service parameters of a network link between two
// Kubernetes nodes that are directly connected to each other.
type NetworkLinkQoS struct {
	// The advertised quality class of this network link
	QualityClass NetworkQualityClass `json:"qualityClass"`

	// The throughput of the network link.
	Throughput NetworkThroughput `json:"throughput"`

	// The latency of the the network link.
	Latency NetworkLatency `json:"latency"`

	// The average packet loss of this network link.
	PacketLoss NetworkPacketLoss `json:"packetLoss"`
}

// NetworkThroughput describes the last known speed of the NetworkLink.
type NetworkThroughput struct {
	// Describes the last known bandwidth of the network link in kilobits per second.
	//
	// +kubebuilder:validation:Minimum=0
	BandwidthKbps int64 `json:"bandwidthKbps"`

	// The variance of BandwidthKbps.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	// +optional
	BandwidthVariance int64 `json:"bandwidthVariance"`
}

// NetworkLatency describes the latency of a NetworkLink.
type NetworkLatency struct {
	// The end-to-end network delay (i.e., latency) of a packet sent between the two nodes, connected by this NetworkLink.
	//
	// +kubebuilder:validation:Minimum=0
	PacketDelayMsec int32 `json:"packetDelayMsec"`

	// The variance of PacketDelayMsec (i.e., jitter).
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=0
	// +optional
	PacketDelayVariance int32 `json:"packetDelayVariance"`
}

// NetworkPacketLoss describes the packet loss of a NetworkLink.
type NetworkPacketLoss struct {
	// The packet loss in basis points (bp).
	// 1 bp = 0.01%
	//
	// The reason for not using percent for this is that the Kubernetes API does not support
	// floating point numbers and people may need more precise packet loss information than whole percents.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10000
	PacketLossBp int32 `json:"packetLossBp"`
}
