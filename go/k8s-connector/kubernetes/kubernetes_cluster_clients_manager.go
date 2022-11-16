package kubernetes

import (
	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

var (
	_ client.ClusterClientsManager = (*KubernetesClusterClientsManager)(nil)
)

// ClusterClientsManager for Kubernetes.
type KubernetesClusterClientsManager struct {
	*client.GenericClusterClientsManager[KubernetesClusterClient]
}

// Creates a new KubernetesClusterClientsManager and initializes it with clients for the specified cluster configurations.
func NewKubernetesClusterClientsManager(clusterConfigs map[string]*rest.Config, parentComponentName string, logger *logr.Logger) (*KubernetesClusterClientsManager, error) {
	clients := make(map[string]KubernetesClusterClient, len(clusterConfigs))

	for clusterName, kubeconfig := range clusterConfigs {
		client, err := NewKubernetesClusterClientImpl(clusterName, kubeconfig, parentComponentName, logger)
		if err != nil {
			return nil, err
		}
		clients[clusterName] = client
	}

	mgr := &KubernetesClusterClientsManager{
		GenericClusterClientsManager: client.NewGenericClusterClientsManager(clients),
	}

	return mgr, nil
}
