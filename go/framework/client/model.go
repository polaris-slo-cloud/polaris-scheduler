package client

import (
	core "k8s.io/api/core/v1"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/util"
)

var (
	_ error = (*PolarisErrorDto)(nil)
)

// Contains the scheduling decision for a pod within a cluster.
type ClusterSchedulingDecision struct {
	// The Pod to be scheduled.
	Pod *core.Pod `json:"pod" yaml:"pod"`

	// The name of the node, to which the pod has been assigned.
	NodeName string `json:"nodeName" yaml:"nodeName"`
}

// Augments a node with information computed by the Polaris Scheduling framework.
type ClusterNode struct {
	*core.Node `json:",inline" yaml:",inline"`

	// The pods that are already scheduled on (bound to) this node.
	Pods []*ClusterPod

	// The pods that are queued to be bound to this node.
	// These pods are currently in the binding pipeline, but their resources are already accounted for in
	// the node's AvailableResources, because committing a scheduling decision may take some time.
	QueuedPods []*ClusterPod

	// The resources that are currently available for allocation on the node.
	//
	// Unlike the Kubernetes Allocatable field, these AvailableResources already accounts for
	// resources consumed by other pods.
	AvailableResources *util.Resources `json:"availableResources" yaml:"availableResources"`

	// The total amount of resources that are available on the node.
	TotalResources *util.Resources `json:"totalResources" yaml:"totalResources"`
}

// Creates a new cluster node, based on the specified node object, assuming that no pods are scheduled on it yet.
func NewClusterNode(node *core.Node) *ClusterNode {
	cn := &ClusterNode{
		Node: node,
		// A new node does not have any pods yet, so both resources are set to the total amount that is allocatable.
		AvailableResources: util.NewResourcesFromList(node.Status.Allocatable),
		TotalResources:     util.NewResourcesFromList(node.Status.Allocatable),
		Pods:               make([]*ClusterPod, 0),
		QueuedPods:         make([]*ClusterPod, 0),
	}
	return cn
}

// Creates a new cluster node, based on the specified node object and the pods that are already scheduled and queued on it.
func NewClusterNodeWithPods(node *core.Node, pods []*ClusterPod, queuedPods []*ClusterPod) *ClusterNode {
	cn := &ClusterNode{
		Node:               node,
		AvailableResources: util.NewResourcesFromList(node.Status.Allocatable),
		TotalResources:     util.NewResourcesFromList(node.Status.Allocatable),
		Pods:               pods,
		QueuedPods:         queuedPods,
	}

	for _, pod := range pods {
		cn.AvailableResources.Subtract(pod.TotalResources)
	}
	for _, pod := range queuedPods {
		cn.AvailableResources.Subtract(pod.TotalResources)
	}

	return cn
}

// Creates a shallow copy of this ClusterNode, i.e., a new object, whose fields point to the same
// objects as the source object.
func (cn *ClusterNode) ShallowCopy() *ClusterNode {
	ret := &ClusterNode{
		Node:               cn.Node,
		Pods:               cn.Pods,
		QueuedPods:         cn.QueuedPods,
		AvailableResources: cn.AvailableResources,
		TotalResources:     cn.TotalResources,
	}
	return ret
}

// Condensed information about an existing pod in a cluster.
type ClusterPod struct {

	// The namespace of the pod.
	Namespace string

	// The name of the pod.
	Name string

	// The total resources consumed by this pod.
	TotalResources *util.Resources

	// Affinity/anti-affinity information.
	// This is nil, if not present on the pod.
	Affinity *core.Affinity
}

// Creates a new ClusterPod, based on the specified pod object.
func NewClusterPod(pod *core.Pod) *ClusterPod {
	cp := &ClusterPod{
		Namespace:      pod.Namespace,
		Name:           pod.Name,
		TotalResources: util.CalculateTotalPodResources(pod),
		Affinity:       pod.Spec.Affinity,
	}
	return cp
}

// Describes timings (in milliseconds) of various phases of the CommitSchedulingDecision operation on the polaris-cluster-agent.
type CommitSchedulingDecisionTimings struct {
	// The time spent in the queue waiting for a binding pipeline to become available.
	QueueTime int64 `json:"queueTime" yaml:"queueTime"`

	// The time spent waiting for the node to be locked for binding.
	NodeLockTime int64 `json:"nodeLockTime" yaml:"nodeLockTime"`

	// The time it takes to fetch the target node and its assigned pods.
	FetchNodeInfo int64 `json:"fetchNodeInfo" yaml:"fetchNodeInfo"`

	// The duration of the binding pipeline
	BindingPipeline int64 `json:"bindingPipeline" yaml:"bindingPipeline"`

	// Commit decision is the entire time it takes to commit a decision using the local cluster client.
	// The commit involves CreatePod, CreateBinding, and any calling overheads.
	CommitDecision int64 `json:"commitDecision" yaml:"commitDecision"`

	// The duration of the request to create a Pod in the orchestrator.
	CreatePod int64 `json:"createPod" yaml:"createPod"`

	// The duration of the request to bind the pod to the target node in the orchestrator.
	CreateBinding int64 `json:"createBinding" yaml:"createBinding"`
}

// Encapsulates the success result of committing a SchedulingDecision.
type CommitSchedulingDecisionSuccess struct {

	// The namespace of the pod.
	Namespace string `json:"namespace" yaml:"namespace"`

	// The name of the pod.
	PodName string `json:"podName" yaml:"podName"`

	// The name of the target node, to which the pod was bound.
	NodeName string `json:"nodeName" yaml:"nodeName"`

	// Timings of the commit operation on the polaris-cluster-agent.
	// Note that when using a LocalClusterClient, only the CreatePod and CreateBinding fields are filled.
	Timings *CommitSchedulingDecisionTimings `json:"timings" yaml:"timings"`
}

// A generic DTO for transmitting error information.
type PolarisErrorDto struct {
	Message string `json:"message" yaml:"message"`
}

func NewPolarisErrorDto(err error) *PolarisErrorDto {
	polarisErr := &PolarisErrorDto{
		Message: err.Error(),
	}
	return polarisErr
}

// Error implements error
func (e *PolarisErrorDto) Error() string {
	return e.Message
}
