package v1

import (
	core "k8s.io/api/core/v1"
)

// ServiceGraphNodeType denotes the type of a ServiceGraphNode.
//
// +kubebuilder:validation:Enum=UserNode;ServiceNode
type ServiceGraphNodeType string

const (
	UserNode    ServiceGraphNodeType = "UserNode"
	ServiceNode ServiceGraphNodeType = "ServiceNode"
)

// ServiceGraphNode describes a node in the ServiceGraph, which may either
// be a service node, which represents a component of the fog application, or
// a user node, which represents the end user of the application.
type ServiceGraphNode struct {

	// Designates the service account used for running the service described by this node
	// and thus, defines the permissions that this service has.
	//
	// +optional
	ServiceAccountName *string `json:"serviceAccountName,omitempty"`

	// The name of this ServiceGraphNode.
	//
	// This must be unique within the graph.
	Name string `json:"name"`

	// Describes whether this node represents an application service or a user.
	// The possible values are: "ServiceNode" (= default) and "UserNode".
	//
	// +kubebuilder:default=ServiceNode
	// +optional
	NodeType ServiceGraphNodeType `json:"nodeType"`

	// The labels that should be applied to the pods, created from this ServiceGraphNode.
	//
	// +optional
	PodLabels map[string]string `json:"labels,omitempty"`

	// Containers that are used to initialize the service upon startup.
	//
	// +optional
	InitContainers []core.Container `json:"initContainers,omitempty"`

	// The containers that constitute the service represented by this node.
	Containers []core.Container `json:"containers"`

	// The storage volumes that should be available to the containers.
	//
	// +optional
	Volumes []core.Volume `json:"volumes,omitempty"`

	// Configures how multiple instances of this node are created.
	Replicas ReplicasConfig `json:"replicas"`

	// Configures if ports should be exposed from this ServiceGraphNode.
	//
	// The exposed ports are available at the DNS name "<ServiceGraphNode.Name>.<ServiceGraphNamespace>.svc"
	// For example, if a ServiceGraph is deployed the namespace "fog" and the ServiceGraphNode's name is "db",
	// the DNS name would be "db.fog.svc"
	//
	// +optional
	ExposedPorts *ExposedPorts `json:"exposedPorts,omitempty"`

	// Allows to configure affinity to cluster nodes and affinity and anti-affinity
	// to other ServiceGraphNodes (referred to as "pods" in the data structure).
	//
	// +optional
	Affinity core.Affinity `json:"affinity,omitempty"`

	// The SLOs defined for this ServiceGraphNode.
	//
	// +optional
	SLOs []ServiceLevelObjective `json:"slos,omitempty"`

	// The set of RAINBOW services that should be available to the instances of this node.
	//
	// +optional
	RainbowServices []RainbowService `json:"rainbowServices,omitempty"`

	// The trust requirements that the hosting cluster node must fulfill.
	//
	// If omitted, the hosting cluster node must not fulfill any trust requirements.
	//
	// +optional
	TrustRequirements *NodeTrustRequirements `json:"trustRequirements,omitempty"`

	// Allows specifying requirements for the cluster node that will host an instance of this ServiceGraphNode.
	//
	// This only specifies what hardware the cluster node needs to have.
	// It does not specify resource requirements or exclusive usage of that hardware.
	//
	// +optional
	NodeHardware *NodeHardware `json:"nodeHardware,omitempty"`

	// Used to constrain the geographical locations where this service may be deployed.
	//
	// +optional
	GeoLocation *GeoLocation `json:"geoLocation,omitempty"`
}
