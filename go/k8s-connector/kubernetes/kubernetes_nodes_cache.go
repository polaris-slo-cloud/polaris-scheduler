package kubernetes

import (
	"context"
	"fmt"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
)

var (
	_ client.NodesCache      = (*KubernetesNodesCache)(nil)
	_ client.PodQueuedOnNode = (*podQueuedOnNode)(nil)
)

// Describes the possible types of cache updates in the updateQueue.
type updateType int

const (
	addition updateType = iota
	update
	removal
)

// Describes a single update for the cache.
// Each update contains either a node or a pod, but never both.
type nodesCacheUpdate struct {
	updateType updateType
	node       *core.Node
	pod        *core.Pod
}

type podQueuedOnNode struct {
	clusterPod *client.ClusterPod
	nativePod  *core.Pod
	nodeName   string
	nodesCache *KubernetesNodesCache
}

// NodesCache implementation for Kubernetes.
type KubernetesNodesCache struct {
	ctx context.Context

	store          collections.ConcurrentObjectStore[*client.ClusterNode]
	updateInterval time.Duration
	updateQueue    chan *nodesCacheUpdate

	clusterClient KubernetesClusterClient
	nodesInformer cache.SharedIndexInformer
	podsInformer  cache.SharedIndexInformer
}

func NewKubernetesNodesCache(
	clusterClient KubernetesClusterClient,
	updateInterval time.Duration,
	queueSize int,
) *KubernetesNodesCache {
	knc := &KubernetesNodesCache{
		store:          collections.NewConcurrentObjectStoreImpl[*client.ClusterNode](),
		clusterClient:  clusterClient,
		updateInterval: updateInterval,
		updateQueue:    make(chan *nodesCacheUpdate, queueSize),
	}

	return knc
}

func (knc *KubernetesNodesCache) Nodes() collections.ConcurrentObjectStore[*client.ClusterNode] {
	return knc.store
}

func (knc *KubernetesNodesCache) QueuePodOnNode(pod *core.Pod, nodeName string) client.PodQueuedOnNode {
	return knc.addQueuedPodToNode(nodeName, pod)
}

func (knc *KubernetesNodesCache) StartWatch(ctx context.Context) error {
	if knc.ctx != nil {
		return fmt.Errorf("watch has already been started")
	}
	knc.ctx = ctx

	factory := informers.NewSharedInformerFactory(knc.clusterClient.ClientSet(), 0)
	knc.nodesInformer = knc.setUpNodesInformer(factory)
	knc.podsInformer = knc.setUpPodsInformer(factory)

	go knc.runUpdateLoop()

	if err := knc.startInformerAndDoInitialSync(knc.nodesInformer); err != nil {
		return err
	}
	return knc.startInformerAndDoInitialSync(knc.podsInformer)
}

func (knc *KubernetesNodesCache) setUpNodesInformer(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
	nodesInformer := factory.Core().V1().Nodes()
	sharedInformer := nodesInformer.Informer()

	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := coerceToNodeOrPanic(obj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: addition,
				node:       node,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			node := coerceToNodeOrPanic(newObj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: update,
				node:       node,
			})
		},
		DeleteFunc: func(obj interface{}) {
			node := coerceToNodeOrPanic(obj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: removal,
				node:       node,
			})
		},
	})

	return sharedInformer
}

func (knc *KubernetesNodesCache) setUpPodsInformer(factory informers.SharedInformerFactory) cache.SharedIndexInformer {
	podsInformer := factory.Core().V1().Pods()
	sharedInformer := podsInformer.Informer()

	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := coerceToPodOrPanic(obj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: addition,
				pod:        pod,
			})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod := coerceToPodOrPanic(newObj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: update,
				pod:        pod,
			})
		},
		DeleteFunc: func(obj interface{}) {
			pod := coerceToPodOrPanic(obj)
			knc.onUpdateArrived(&nodesCacheUpdate{
				updateType: removal,
				pod:        pod,
			})
		},
	})

	return sharedInformer
}

func (knc *KubernetesNodesCache) startInformerAndDoInitialSync(informer cache.SharedIndexInformer) error {
	go informer.Run(knc.ctx.Done())

	if !cache.WaitForCacheSync(knc.ctx.Done(), informer.HasSynced) {
		err := fmt.Errorf("timed out waiting for caches to sync")
		runtime.HandleError(err)
		return err
	}

	return nil
}

func (knc *KubernetesNodesCache) onUpdateArrived(update *nodesCacheUpdate) {
	knc.updateQueue <- update
}

func (knc *KubernetesNodesCache) runUpdateLoop() {
	stopCh := knc.ctx.Done()

	for {
		select {
		case <-stopCh:
			return
		default:
			knc.runUpdateLoopIteration()
			time.Sleep(knc.updateInterval)
		}
	}
}

// Processes all updates currently in the queue.
func (knc *KubernetesNodesCache) runUpdateLoopIteration() {
	var storeWriter collections.ConcurrentObjectStoreWriter[*client.ClusterNode]

	for {
		select {
		case update := <-knc.updateQueue:
			if storeWriter == nil {
				storeWriter = knc.store.WriteLock()
				defer storeWriter.Unlock()
			}
			if update.node != nil {
				knc.applyNodeUpdate(update, storeWriter)
			} else {
				knc.applyPodUpdate(update, storeWriter)
			}
		default:
			return
		}
	}
}

func (knc *KubernetesNodesCache) applyNodeUpdate(cacheUpdate *nodesCacheUpdate, storeWriter collections.ConcurrentObjectStoreWriter[*client.ClusterNode]) {
	switch cacheUpdate.updateType {
	case addition:
		clusterNode := client.NewClusterNode(cacheUpdate.node)
		storeWriter.Set(cacheUpdate.node.Name, clusterNode)
	case removal:
		storeWriter.Remove(cacheUpdate.node.Name)
	case update:
		// This seems to happen quite frequently in Kubernetes, probably due to heartbeats from the nodes.
		clusterNode := knc.createUpdatedClusterNode(cacheUpdate.node, storeWriter)
		storeWriter.Set(cacheUpdate.node.Name, clusterNode)
	}
}

// Creates an updated version of a ClusterNode that already exists in the cache.
func (knc *KubernetesNodesCache) createUpdatedClusterNode(updatedNode *core.Node, storeReader collections.ConcurrentObjectStoreWriter[*client.ClusterNode]) *client.ClusterNode {
	var updatedClusterNode *client.ClusterNode

	if oldClusterNode, ok := storeReader.GetByKey(updatedNode.Name); ok {
		// No need to copy the pods slices, because the ClusterNodes should be treated as immutable anyway.
		updatedClusterNode = client.NewClusterNodeWithPods(updatedNode, oldClusterNode.Pods, oldClusterNode.QueuedPods)
	} else {
		updatedClusterNode = client.NewClusterNode(updatedNode)
	}

	return updatedClusterNode
}

func (knc *KubernetesNodesCache) applyPodUpdate(cacheUpdate *nodesCacheUpdate, storeWriter collections.ConcurrentObjectStoreWriter[*client.ClusterNode]) {
	var updatedClusterNode *client.ClusterNode
	oldClusterNode, ok := storeWriter.GetByKey(cacheUpdate.pod.Spec.NodeName)
	if !ok {
		// Node does not exist (could be a user error, because this field can be set by the user), so we ignore the pod.
		return
	}

	switch cacheUpdate.updateType {
	case addition:
		updatedClusterNode = knc.addOrUpdatePodOnNode(oldClusterNode, cacheUpdate.pod, false)
	case removal:
		updatedClusterNode = knc.removePodFromNode(oldClusterNode, cacheUpdate.pod)
	case update:
		updatedClusterNode = knc.addOrUpdatePodOnNode(oldClusterNode, cacheUpdate.pod, false)
	}

	storeWriter.Set(updatedClusterNode.Name, updatedClusterNode)
}

// Adds or updates the specified pod to/on a copy of the clusterNode and returns that updated copy.
//
// We unify addition and update in one method, because if the pod was previously a PodQueuedOnNode and marked as committed,
// it will already be present in the node's list of pods.
// If removeFromQueuedPods is true, the pod will also be removed from the QueuedPods list.
func (knc *KubernetesNodesCache) addOrUpdatePodOnNode(oldNode *client.ClusterNode, pod *core.Pod, removeFromQueuedPods bool) *client.ClusterNode {
	clusterPod := client.NewClusterPod(pod)

	newPods := make([]*client.ClusterPod, len(oldNode.Pods), len(oldNode.Pods)+1)
	podUpdated := copyPodsAndUpdateItem(newPods, oldNode.Pods, clusterPod)
	if !podUpdated {
		newPods = append(newPods, clusterPod)
	}

	newQueuedPods := oldNode.QueuedPods
	if removeFromQueuedPods {
		newQueuedPods = make([]*client.ClusterPod, len(oldNode.QueuedPods)-1)
		copyPodsAndRemoveItem(newQueuedPods, oldNode.QueuedPods, pod)
	}

	return client.NewClusterNodeWithPods(oldNode.Node, newPods, newQueuedPods)
}

// Removes the specified pod to a copy of the clusterNode and returns that updated copy.
func (knc *KubernetesNodesCache) removePodFromNode(oldNode *client.ClusterNode, podToRemove *core.Pod) *client.ClusterNode {
	updatedNode := &client.ClusterNode{
		Node:               oldNode.Node,
		AvailableResources: oldNode.AvailableResources.DeepCopy(),
		TotalResources:     oldNode.TotalResources.DeepCopy(),
		Pods:               make([]*client.ClusterPod, len(oldNode.Pods)-1),
		QueuedPods:         oldNode.QueuedPods,
	}

	removedPod := copyPodsAndRemoveItem(updatedNode.Pods, oldNode.Pods, podToRemove)
	updatedNode.AvailableResources.Add(removedPod.TotalResources)
	return updatedNode
}

// Adds the specified pod to the node's queue and updates its resources.
// To avoid cache staleness when the number of scheduling commits is high, we execute this operation immediately and do not queue it.
func (knc *KubernetesNodesCache) addQueuedPodToNode(nodeName string, pod *core.Pod) *podQueuedOnNode {
	storeWriter := knc.store.WriteLock()
	defer storeWriter.Unlock()

	oldNode, ok := storeWriter.GetByKey(nodeName)
	if !ok {
		return nil
	}

	updatedNode := &client.ClusterNode{
		Node:               oldNode.Node,
		AvailableResources: oldNode.AvailableResources.DeepCopy(),
		TotalResources:     oldNode.TotalResources.DeepCopy(),
		Pods:               oldNode.Pods,
		QueuedPods:         make([]*client.ClusterPod, len(oldNode.QueuedPods), len(oldNode.QueuedPods)+1),
	}
	copy(updatedNode.QueuedPods, oldNode.QueuedPods)

	clusterPod := client.NewClusterPod(pod)
	updatedNode.QueuedPods = append(updatedNode.QueuedPods, clusterPod)
	updatedNode.AvailableResources.Subtract(clusterPod.TotalResources)

	storeWriter.Set(nodeName, updatedNode)
	return knc.createPodQueuedOnNode(clusterPod, pod, nodeName)
}

func (knc *KubernetesNodesCache) createPodQueuedOnNode(clusterPod *client.ClusterPod, nativePod *core.Pod, nodeName string) *podQueuedOnNode {
	queuedPod := &podQueuedOnNode{
		clusterPod: clusterPod,
		nativePod:  nativePod,
		nodeName:   nodeName,
		nodesCache: knc,
	}
	return queuedPod
}

func coerceToNodeOrPanic(obj interface{}) *core.Node {
	node, ok := obj.(*core.Node)
	if !ok {
		panic("NodeInformer received a non Node object")
	}
	return node
}

func coerceToPodOrPanic(obj interface{}) *core.Pod {
	node, ok := obj.(*core.Pod)
	if !ok {
		panic("PodInformer received a non Pod object")
	}
	return node
}

// Copies the pods from src to dest, removing the podToRemove.
// Returns the ClusterPod that was removed (not copied) or nil, if this pod could not be found.
func copyPodsAndRemoveItem(dest []*client.ClusterPod, src []*client.ClusterPod, podToRemove *core.Pod) *client.ClusterPod {
	var removedPod *client.ClusterPod
	destIndex := 0
	for _, currPod := range src {
		if currPod.Name != podToRemove.Name || currPod.Namespace != podToRemove.Namespace {
			dest[destIndex] = currPod
			destIndex++
		} else {
			removedPod = currPod
		}
	}
	return removedPod
}

// Copies the pods from src to dest, replacing the pod, whose name and namespace matches podToUpdate with podToUpdate.
// Returns true, if the pod was updated, or false if this pod could not be found.
func copyPodsAndUpdateItem(dest []*client.ClusterPod, src []*client.ClusterPod, podToUpdate *client.ClusterPod) bool {
	podUpdated := false
	for i, currPod := range src {
		if currPod.Name != podToUpdate.Name || currPod.Namespace != podToUpdate.Namespace {
			dest[i] = currPod
		} else {
			dest[i] = podToUpdate
			podUpdated = true
		}
	}
	return podUpdated
}

func (p *podQueuedOnNode) NodeName() string {
	return p.nodeName
}

func (p *podQueuedOnNode) Pod() *client.ClusterPod {
	return p.clusterPod
}

func (p *podQueuedOnNode) MarkAsCommitted() {
	storeWriter := p.nodesCache.store.WriteLock()
	defer storeWriter.Unlock()

	oldNode, ok := storeWriter.GetByKey(p.nodeName)
	if !ok {
		return
	}

	updatedNode := p.nodesCache.addOrUpdatePodOnNode(oldNode, p.nativePod, true)
	storeWriter.Set(p.nodeName, updatedNode)
}

func (p *podQueuedOnNode) RemoveFromQueue() {
	storeWriter := p.nodesCache.store.WriteLock()
	defer storeWriter.Unlock()

	oldNode, ok := storeWriter.GetByKey(p.nodeName)
	if !ok {
		return
	}

	updatedNode := &client.ClusterNode{
		Node:               oldNode.Node,
		AvailableResources: oldNode.AvailableResources.DeepCopy(),
		TotalResources:     oldNode.TotalResources.DeepCopy(),
		Pods:               oldNode.Pods,
		QueuedPods:         make([]*client.ClusterPod, len(oldNode.QueuedPods)-1),
	}

	removedPod := copyPodsAndRemoveItem(updatedNode.QueuedPods, oldNode.QueuedPods, p.nativePod)
	updatedNode.AvailableResources.Add(removedPod.TotalResources)
	storeWriter.Set(p.nodeName, updatedNode)
}
