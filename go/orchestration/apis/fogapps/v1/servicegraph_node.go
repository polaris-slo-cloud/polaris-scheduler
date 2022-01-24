package v1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// References to secrets for pulling container images from private registries.
	//
	// +optional
	ImagePullSecrets []core.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// The storage volumes that should be available to the containers.
	//
	// +optional
	Volumes []core.Volume `json:"volumes,omitempty"`

	// Configures how multiple instances of this node are created.
	Replicas ReplicasConfig `json:"replicas"`

	// If true, a pod created from this ServiceGraphNode, will use host node's network namespace.
	//
	// +kubebuilder:default=false
	// +optional
	HostNetwork bool `json:"hostNetwork,omitempty"`

	// Configures if ports should be exposed from this ServiceGraphNode.
	//
	// The exposed ports are available at the DNS name "<ServiceGraphNode.Name>.<ServiceGraphNamespace>.svc"
	// For example, if a ServiceGraph is deployed the namespace "fog" and the ServiceGraphNode's name is "db",
	// the DNS name would be "db.fog.svc". For other services within the same ServiceGraph "<ServiceGraphNode.Name>"
	// is enough (e.g., "db").
	//
	// +optional
	ExposedPorts *ExposedPorts `json:"exposedPorts,omitempty"`

	// Allows to configure affinity to cluster nodes and affinity and anti-affinity
	// to other ServiceGraphNodes (referred to as "pods" in the data structure).
	//
	// +optional
	Affinity *core.Affinity `json:"affinity,omitempty"`

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

// ServiceGraphNodeStatus describes the observed state of the resources created from a ServiceGraphNode.
type ServiceGraphNodeStatus struct {

	// Describes the type of deployment resource that is created for this ServiceGraphNode.
	//
	// +optional
	DeploymentResourceType *metav1.GroupVersionKind `json:"deploymentType,omitempty"`

	// The last observed initial replicas value observed on the ServiceGraphNode.
	// If the value on the ServiceGraphNode is different, the number of replicas on the deployment
	// is changed (and a value potentially set through an elasticity strategy is overwritten),
	// otherwise we know that the initial number of replicas has not been changed on
	// the ServiceGraph and we do not need to update the deployment.
	//
	// +optional
	InitialReplicas int32 `json:"initialReplicas,omitempty"`

	// The number of replicas that has been configured, based on the ServiceGraphNode and the
	// state of the SLOs.
	//
	// +optional
	ConfiguredReplicas int32 `json:"configuredReplicas,omitempty"`

	// The number of replicas in the Ready state that have been observed.
	//
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}
