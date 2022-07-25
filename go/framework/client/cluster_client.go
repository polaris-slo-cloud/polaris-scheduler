package client

import (
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/events"
)

// Represents a client for communicating with a single cluster.
type ClusterClient interface {
	// Gets the ClientSet for communicating with a Kubernetes cluster.
	ClientSet() clientset.Interface

	// Gets the EventRecorder for this cluster.
	EventRecorder() events.EventRecorder
}

// Manages the clients for multiple clusters.
type ClusterClientsManager interface {

	// Gets the ClusterClient for the specified cluster.
	GetClusterClient(clusterName string) (ClusterClient, error)
}
