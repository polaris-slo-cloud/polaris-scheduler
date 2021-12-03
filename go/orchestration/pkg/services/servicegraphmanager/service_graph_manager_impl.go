package servicegraphmanager

import (
	"context"
	"fmt"
	"sync"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_ ServiceGraphManager = (*serviceGraphManagerImpl)(nil)
)

type serviceGraphID struct {
	types.NamespacedName

	// <namespace>.<name>
	mapKey string
}

type serviceGraphManagerImpl struct {
	// Maps a service graph identifier (<namespace>.<name>) to a Future, whose
	// result is a serviceGraphStateImpl.
	activeStates sync.Map

	// The K8s client used to load ServiceGraph CRD instances.
	client client.Client
}

func newServiceGraphManagerImpl(k8sClient client.Client) *serviceGraphManagerImpl {
	return &serviceGraphManagerImpl{
		activeStates: sync.Map{},
		client:       k8sClient,
	}
}

func (me *serviceGraphManagerImpl) AcquireServiceGraphState(pod *core.Pod) (ServiceGraphState, error) {
	if svcGraphID := me.getServiceGraphID(pod); svcGraphID != nil {
		return me.acquireServiceGraphStateInternal(svcGraphID, pod)
	}
	return nil, nil
}

func (me *serviceGraphManagerImpl) acquireServiceGraphStateInternal(svcGraphID *serviceGraphID, pod *core.Pod) (ServiceGraphState, error) {
	var svcGraphState *serviceGraphStateImpl
	var err error

	// We use a loop to get and acquire the ServiceGraphState, because we might get a state from the map
	// and then the last goroutine using it releases it and marks it for deletion, before we can acquire it.
	// In such a case acquire() will return false and we need to try again (this time we will likely create a new ServiceGraphState).
	for acquired := false; !acquired; {
		if svcGraphState, err = me.getOrCreateServiceGraphState(svcGraphID); err != nil {
			return nil, err
		}
		acquired = svcGraphState.acquire(pod)
	}
	return svcGraphState, err
}

// Returns the service graph identifier from the pod
// or nil if the pod does not have any service graph info attached.
func (me *serviceGraphManagerImpl) getServiceGraphID(pod *core.Pod) *serviceGraphID {
	if svcGraphName, ok := kubeutil.GetLabel(pod, kubeutil.LabelRefServiceGraph); ok {
		namespace := kubeutil.GetNamespace(pod)
		return &serviceGraphID{
			NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      svcGraphName,
			},
			mapKey: me.getServiceGraphMapKey(namespace, svcGraphName),
		}
	}
	return nil
}

// Returns the map key for the ServiceGraph with the given namespace and name.
func (me *serviceGraphManagerImpl) getServiceGraphMapKey(namespace, svcGraphName string) string {
	return fmt.Sprintf("%s.%s", namespace, svcGraphName)
}

// Gets a ServiceGraphState handle from the activeStates map and resolves the handle to the state object or
// creates a new handle and state object.
func (me *serviceGraphManagerImpl) getOrCreateServiceGraphState(svcGraphID *serviceGraphID) (*serviceGraphStateImpl, error) {
	var handle util.Future

	if existingHandle, ok := me.activeStates.Load(svcGraphID.mapKey); ok {
		handle = existingHandle.(util.Future)
	} else {
		// Create a new Future handle and try storing it in the map.
		// If another goroutine has stored a handle in the meantime, we use that one instead.
		// If our Future handle was stored, we load the ServiceGraphState.
		newHandle, resultProvider := util.NewFuture()
		if actualHandle, loaded := me.activeStates.LoadOrStore(svcGraphID.mapKey, newHandle); loaded {
			handle = actualHandle.(util.Future)
		} else {
			handle = newHandle
			me.createServiceGraphState(svcGraphID, resultProvider)
		}
	}

	svcGraphState, err := handle.Get()
	return svcGraphState.(*serviceGraphStateImpl), err
}

// Loads the ServiceGraph CRD, assembles the respective graph, and starts asynchronously loading the placement map.
func (me *serviceGraphManagerImpl) createServiceGraphState(svcGraphID *serviceGraphID, resultProvider util.ResultProvider) {
	var crd fogappsCRDs.ServiceGraph
	if err := me.client.Get(context.TODO(), svcGraphID.NamespacedName, &crd); err != nil {
		resultProvider(nil, err)
	}

	graph := servicegraph.FromCRDInstance(&crd)
	placementMapHandle, placementMapResultProvider := util.NewFuture()

	svcGraphState := newServiceGraphStateImpl(
		&graph,
		&crd,
		placementMapHandle,
		func(state *serviceGraphStateImpl) { me.deleteServiceGraphState(state) },
	)
	resultProvider(svcGraphState, nil)
	go me.buildPlacementMap(svcGraphState, placementMapResultProvider)
}

func (me *serviceGraphManagerImpl) deleteServiceGraphState(svcGraphState ServiceGraphState) {
	crd := svcGraphState.ServiceGraphCRD()
	mapKey := me.getServiceGraphMapKey(crd.Namespace, crd.Name)
	me.activeStates.Delete(mapKey)
}

func (me *serviceGraphManagerImpl) buildPlacementMap(svcGraphState ServiceGraphState, resultProvider util.ResultProvider) {

}
