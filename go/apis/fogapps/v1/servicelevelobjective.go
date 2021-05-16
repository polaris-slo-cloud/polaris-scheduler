package v1

// ServiceLevelObjective an SLOs that is attached to
// a ServiceGraph, a ServiceGraphNode, or a ServiceLink.
type ServiceLevelObjective struct {

	// Describes the type of SLO.
	SloType ApiVersionKind `json:"sloType"`

	// The name of this SLO instance.
	//
	// This must be unique within its containing list of SLOs (e.g., ServiceGraphNode.SLOs if this SLO is attached to a node).
	Name string `json:"name"`

	// The elasticity strategy that should be triggered upon violations of the SLO.
	ElasticityStrategy ApiVersionKind `json:"elasticityStrategy"`

	// The SLO-specific configuration.
	Config *ArbitraryObject `json:"config,omitempty"`
}
