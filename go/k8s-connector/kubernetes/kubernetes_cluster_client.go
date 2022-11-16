package kubernetes

import (
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

// Kubernetes-specific extension of the ClusterClient.
type KubernetesClusterClient interface {
	client.ClusterClient

	// Gets the ClientSet for communicating with a Kubernetes cluster.
	ClientSet() clientset.Interface

	// Gets the EventRecorder for this cluster.
	EventRecorder() record.EventRecorder
}
