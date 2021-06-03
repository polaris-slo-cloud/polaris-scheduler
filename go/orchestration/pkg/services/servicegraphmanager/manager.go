package servicegraphmanager

import (
	v1 "k8s.io/api/core/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

var (
	instance ServiceGraphManager
)

// ServiceGraphManager provides methods for obtaining the service graph for a particular application
type ServiceGraphManager interface {
	// ServiceGraph gets the service graph for the application that the specified pod is part of.
	ServiceGraph(pod *v1.Pod) (*servicegraph.ServiceGraph, error)
}

// GetServiceGraphManager returns the singleton instance of the ServiceGraphManager.
func GetServiceGraphManager() ServiceGraphManager {
	if instance == nil {
		instance = newServiceGraphManagerImpl()
	}
	return instance
}
