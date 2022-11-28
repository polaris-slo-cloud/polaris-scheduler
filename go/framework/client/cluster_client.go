package client

import (
	"context"

	core "k8s.io/api/core/v1"
)

// Represents a client for communicating with a single cluster.
type ClusterClient interface {
	// Gets the name of the cluster.
	ClusterName() string

	// Commits the scheduling decision to the cluster.
	CommitSchedulingDecision(ctx context.Context, schedulingDecision *ClusterSchedulingDecision) error

	// ToDo: Add generic methods for accessing arbitrary cluster objects? or at least a defined subset?

	// Originally, we considered using a native only client in the cluster agent, but then we decided to
	// keep this abstraction for the cluster agent, because it allows orchestrator-independent cluster agent plugins.
}

// A superset of ClusterClient with more capabilities and which is only available in the ClusterAgent.
type LocalClusterClient interface {
	ClusterClient

	// Fetches the node with the specified name.
	FetchNode(ctx context.Context, name string) (*core.Node, error)

	// Fetches all pods that are currently scheduled on the node with the specified name.
	//
	// ToDo: Refactor this into a fetch method with more generic search criteria and maybe return an array of pointers.
	FetchPodsScheduledOnNode(ctx context.Context, nodeName string) ([]core.Pod, error)
}

// Manages the clients for multiple clusters.
type ClusterClientsManager interface {

	// Gets the ClusterClient for the specified cluster or an error, if the specified cluster cannot be found.
	GetClusterClient(clusterName string) (ClusterClient, error)

	// Gets the number of clusters known to this ClusterClientsManager.
	ClustersCount() int

	// Calls the specified function for each cluster.
	// If the function returns an error, the iteration is stopped immediately and returns that error.
	ForEach(fn func(clusterName string, clusterClient ClusterClient) error) error
}
