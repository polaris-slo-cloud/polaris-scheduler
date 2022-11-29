package client

import (
	"context"

	core "k8s.io/api/core/v1"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
)

// Tracks and caches the full list of nodes in a cluster and their available resources.
//
// The node objects in the cache are always the most recent ones received through the watch.
//
// As pods are assigned to nodes by the scheduler, it is recommended to use the QueuePodOnNode() method
// to update the node's resources in the cache before committing the scheduling decision to the cluster,
// as this may take some time, during which the node's status in the cache would be outdated.
type NodesCache interface {
	// Starts watching the nodes.
	// This method returns once the initial list has been retrieved and added to the store.
	//
	// The passed context can be used to stop the watch.
	StartWatch(ctx context.Context) error

	// Gets the cache of all nodes.
	Nodes() collections.ConcurrentObjectStore[*ClusterNode]

	// Adds the pod as to be bound to the specified node and updates the node's available resources.
	//
	// Use the returned PodQueuedOnNode to inform the cache of the outcome of the commit scheduling decision operation.
	QueuePodOnNode(pod *core.Pod, nodeName string) PodQueuedOnNode
}

// Handle to a pod that was queued to be bound to a node.
// This object is returned by NodesCache.QueuePodOnNode() and must be used
// to inform the cache of the final result of the commit operation.
type PodQueuedOnNode interface {

	// The pod that was queued (immutable).
	Pod() *ClusterPod

	// The name of the node, on which the pod was queued.
	NodeName() string

	// Removes this pod from the node's queue, without marking it as committed,
	// and updates the node's resources.
	RemoveFromQueue()

	// Marks the pod as committed, i.e., moves it from the queue to the list of pods running on the node.
	MarkAsCommitted()
}
