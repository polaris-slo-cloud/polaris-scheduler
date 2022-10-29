package servicegraphmanager

import (
	core "k8s.io/api/core/v1"
)

var (
	instance ServiceGraphManager
)

// We need an AcquireServiceGraphState method that
// 1. Loads the ServiceGraph, if it has not been loaded yet and starts fetching the placement map (creates a ServiceGraphState object).
// 2. Increments the reference count on the service graph state object.

// ServiceGraphManager provides methods for obtaining the service graph for a particular application
type ServiceGraphManager interface {
	// Gets the ServiceGraphState for the application that the specified pod is part of.
	// If the pod has no ServiceGraph associated, the ServiceGraphState AND the error will be nil.
	// Note that this method only checks the pod's reference to the ServiceGraph, but not the reference
	// to a node within that ServiceGraph.
	//
	// The requesting pod is added to reference count of the ServiceGraphState. When the pod's scheduling cycle ends,
	// the state's Release() method must be called to allow it to be removed from the cache. Each pod may call
	// AcquireServiceGraphState() multiple times, it is only counted once towards the reference count.
	//
	// If the ServiceGraph has already been loaded for another pod that is currently in the pipeline, it is returned immediately.
	// If not, the ServiceGraph CRD is fetched (blocking the caller until this has completed) and then the building of the
	// ServiceGraphPlacementMap is started asynchronously.
	AcquireServiceGraphState(pod *core.Pod) (ServiceGraphState, error)
}

// GetServiceGraphManager returns the singleton instance of the ServiceGraphManager.
func GetServiceGraphManager() ServiceGraphManager {
	if instance != nil {
		return instance
	}
	panic("ServiceGraphManager singleton has not been initialized. Did you call InitServiceGraphManager()?")
}

// Initializes the singleton instance of the ServiceGraphManager.
func InitServiceGraphManager() ServiceGraphManager {
	instance = newServiceGraphManagerImpl()
	return instance
}
