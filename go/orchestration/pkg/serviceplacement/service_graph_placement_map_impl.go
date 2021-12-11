package serviceplacement

import (
	"sync"
)

var (
	_servicePlacementMapImpl *serviceGraphPlacementMapImpl

	_ ServiceGraphPlacementMap = _servicePlacementMapImpl
)

//////////////////////////////////////////////////////
// servicePlacementInfo
//////////////////////////////////////////////////////

// Stores placement information for one ServiceGraphNode
type servicePlacementInfo struct {
	// The name of the ServiceGraphNode
	svcGraphNodeLabel string

	// Used for controlling access to k8sNodes
	mutex sync.RWMutex

	// The list of Kubernetes nodes, the pods of this ServiceGraphNode have been placed on.
	k8sNodes []string
}

func newSericePlacementInfo(svcGraphNodeLabel string) *servicePlacementInfo {
	return &servicePlacementInfo{
		svcGraphNodeLabel: svcGraphNodeLabel,
		mutex:             sync.RWMutex{},
		k8sNodes:          make([]string, 0),
	}
}

func (me *servicePlacementInfo) getK8sNodes() []string {
	me.mutex.RLock()
	defer me.mutex.RUnlock()
	ret := me.k8sNodes
	return ret
}

func (me *servicePlacementInfo) updateK8sNodes(updateFn StringSliceTransformFn) {
	me.mutex.Lock()
	defer me.mutex.Unlock()
	me.k8sNodes = updateFn(me.k8sNodes)
}

//////////////////////////////////////////////////////
// serviceGraphPlacementMapImpl
//////////////////////////////////////////////////////

// Default implementation of ServiceGraphPlacementMap
type serviceGraphPlacementMapImpl struct {
	// Controls access to svcGraphNodes
	mutex sync.RWMutex

	// Maps ServiceGraphNode labels to their placement infos
	svcGraphNodes map[string]*servicePlacementInfo

	// Indicates if this is the initial placement of the ServiceGraph
	isInitialPlacement bool
}

func newServicePlacementMapImpl(isInitialPlacement bool) *serviceGraphPlacementMapImpl {
	ret := serviceGraphPlacementMapImpl{
		mutex:              sync.RWMutex{},
		svcGraphNodes:      make(map[string]*servicePlacementInfo),
		isInitialPlacement: isInitialPlacement,
	}
	return &ret
}

func (me *serviceGraphPlacementMapImpl) GetKubernetesNodes(svcGraphNodeLabel string) []string {
	me.mutex.RLock()
	placementInfo, found := me.svcGraphNodes[svcGraphNodeLabel]
	me.mutex.RUnlock()

	if found {
		return placementInfo.getK8sNodes()
	}
	return nil
}

func (me *serviceGraphPlacementMapImpl) SetKubernetesNodes(svcGraphNodeLabel string, updateFn StringSliceTransformFn) {
	placementInfo := me.getOrCreatePlacementInfo(svcGraphNodeLabel)
	placementInfo.updateK8sNodes(updateFn)
}

func (me *serviceGraphPlacementMapImpl) IsInitialPlacement() bool {
	return me.isInitialPlacement
}

func (me *serviceGraphPlacementMapImpl) getOrCreatePlacementInfo(svcGraphNodeLabel string) *servicePlacementInfo {
	me.mutex.Lock()
	defer me.mutex.Unlock()

	placementInfo, found := me.svcGraphNodes[svcGraphNodeLabel]
	if !found {
		placementInfo = newSericePlacementInfo(svcGraphNodeLabel)
		me.svcGraphNodes[svcGraphNodeLabel] = placementInfo
	}
	return placementInfo
}
