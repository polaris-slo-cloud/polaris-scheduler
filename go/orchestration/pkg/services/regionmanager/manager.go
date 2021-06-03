package regionmanager

import (
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/regiongraph"
)

var (
	instance RegionManager
)

// RegionManager provides methods for obtaining information about a fog region
type RegionManager interface {
	// RegionGraph gets a graph that represents that current state of the region.
	RegionGraph() *regiongraph.RegionGraph
}

// GetRegionManager returns the singleton instance of the RegionManager.
func GetRegionManager() RegionManager {
	if instance == nil {
		instance = newRegionManagerImpl()
	}
	return instance
}
