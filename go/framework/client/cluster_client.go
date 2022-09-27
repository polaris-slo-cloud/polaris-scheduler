package client

import (
	"context"
)

// Represents a client for communicating with a single cluster.
type ClusterClient interface {
	// Gets the name of the cluster.
	ClusterName() string

	// Commits the scheduling decision to the cluster.
	CommitSchedulingDecision(ctx context.Context, schedulingDecision *ClusterSchedulingDecision) error

	// ToDo: Add generic methods for accessing arbitrary cluster objects? or at least a defined subset?
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
