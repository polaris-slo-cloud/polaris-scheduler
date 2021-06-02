package v1

// ServiceLink is used to describe runtime relationships between two nodes in a ServiceGraph.
//
// A ServiceLink from nodeA to nodeB indicates that at some point(s) during the lifetime of the application,
// service A makes a request to service B.
type ServiceLink struct {

	// The name of the source node.
	Source string `json:"source"`

	// The name of the target node.
	Target string `json:"target"`

	// The QoS requirements that the underlying NetworkLink must fulfill.
	//
	// +optional
	QosRequirements *LinkQosRequirements `json:"qosRequirements,omitempty"`

	// The trust requirements that the underlying NetworkLink must fulfill.
	//
	// If omitted, the NetworkLink must not fulfill any trust requirements.
	//
	// +optional
	TrustRequirements *LinkTrustRequirements `json:"trustRequirements,omitempty"`

	// The SLOs defined for this link.
	//
	// +optional
	SLOs []ServiceLevelObjective `json:"slos,omitempty"`
}
