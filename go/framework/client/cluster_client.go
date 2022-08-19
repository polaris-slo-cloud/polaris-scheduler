package client

import (
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
)

// Represents a client for communicating with a single cluster.
type ClusterClient interface {
	// Gets the ClientSet for communicating with a Kubernetes cluster.
	ClientSet() clientset.Interface

	// Gets the EventRecorder for this cluster.
	EventRecorder() record.EventRecorder
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
