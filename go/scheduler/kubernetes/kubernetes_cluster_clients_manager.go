package kubernetes

import (
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

var (
	_ client.ClusterClientsManager = (*KubernetesClusterClientsManager)(nil)
)

// ClusterClientsManager for Kubernetes.
type KubernetesClusterClientsManager struct {
	clients map[string]client.ClusterClient
}

// Creates a new KubernetesClusterClientsManager and initializes it with clients for the specified cluster configurations.
func NewKubernetesClusterClientsManager(clusterConfigs map[string]*rest.Config, schedConfig *config.SchedulerConfig, logger *logr.Logger) (*KubernetesClusterClientsManager, error) {
	mgr := KubernetesClusterClientsManager{
		clients: make(map[string]client.ClusterClient, len(clusterConfigs)),
	}

	for clusterName, kubeconfig := range clusterConfigs {
		client, err := NewKubernetesClusterClient(kubeconfig, schedConfig, logger)
		if err != nil {
			return nil, err
		}
		mgr.clients[clusterName] = client
	}

	return &mgr, nil
}

func (mgr *KubernetesClusterClientsManager) GetClusterClient(clusterName string) (client.ClusterClient, error) {
	if client, ok := mgr.clients[clusterName]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("could not find a ClusterClient for cluster %s", clusterName)
}

func (mgr *KubernetesClusterClientsManager) ClustersCount() int {
	return len(mgr.clients)
}

func (mgr *KubernetesClusterClientsManager) ForEach(fn func(clusterName string, client client.ClusterClient) error) error {
	for cluster, client := range mgr.clients {
		if err := fn(cluster, client); err != nil {
			return err
		}
	}
	return nil
}
