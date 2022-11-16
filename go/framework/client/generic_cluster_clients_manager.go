package client

import (
	"fmt"
)

var (
	_ ClusterClientsManager = (*GenericClusterClientsManager[ClusterClient])(nil)
)

// A generic default implementation of the ClusterClientsManager.
type GenericClusterClientsManager[T ClusterClient] struct {
	clients map[string]T
}

func NewGenericClusterClientsManager[T ClusterClient](clients map[string]T) *GenericClusterClientsManager[T] {
	mgr := &GenericClusterClientsManager[T]{
		clients: clients,
	}
	return mgr
}

func (mgr *GenericClusterClientsManager[T]) ClustersCount() int {
	return len(mgr.clients)
}

func (mgr *GenericClusterClientsManager[T]) ForEach(fn func(clusterName string, clusterClient ClusterClient) error) error {
	for cluster, client := range mgr.clients {
		if err := fn(cluster, client); err != nil {
			return err
		}
	}
	return nil
}

func (mgr *GenericClusterClientsManager[T]) GetClusterClient(clusterName string) (ClusterClient, error) {
	if client, ok := mgr.clients[clusterName]; ok {
		return client, nil
	}
	return nil, fmt.Errorf("could not find a ClusterClient for cluster %s", clusterName)
}
