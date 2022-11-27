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
	_ client.NodesCache = (*KubernetesNodesCache)(nil)
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
		// No need to copy the pods slice, because the ClusterNodes should be treated as immutable anyway.
		updatedClusterNode = client.NewClusterNodeWithPods(updatedNode, oldClusterNode.Pods)
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
		updatedClusterNode = knc.addPodToNode(oldClusterNode, cacheUpdate.pod)
	case removal:
		updatedClusterNode = knc.removePodFromNode(oldClusterNode, cacheUpdate.pod)
	case update:
		updatedClusterNode = knc.updatePodOnNode(oldClusterNode, cacheUpdate.pod)
	}

	storeWriter.Set(updatedClusterNode.Name, updatedClusterNode)
}

// Adds the specified pod to a copy of the clusterNode and returns that updated copy.
func (knc *KubernetesNodesCache) addPodToNode(oldNode *client.ClusterNode, pod *core.Pod) *client.ClusterNode {
	updatedNode := &client.ClusterNode{
		Node:               oldNode.Node,
		AvailableResources: oldNode.AvailableResources.DeepCopy(),
		TotalResources:     oldNode.TotalResources.DeepCopy(),
		Pods:               make([]*client.ClusterPod, len(oldNode.Pods), len(oldNode.Pods)+1),
	}
	copy(updatedNode.Pods, oldNode.Pods)

	clusterPod := client.NewClusterPod(pod)
	updatedNode.Pods = append(updatedNode.Pods, clusterPod)
	updatedNode.AvailableResources.Subtract(clusterPod.TotalResources)

	return updatedNode
}

// Removes the specified pod to a copy of the clusterNode and returns that updated copy.
func (knc *KubernetesNodesCache) removePodFromNode(oldNode *client.ClusterNode, podToRemove *core.Pod) *client.ClusterNode {
	updatedNode := &client.ClusterNode{
		Node:               oldNode.Node,
		AvailableResources: oldNode.AvailableResources.DeepCopy(),
		TotalResources:     oldNode.TotalResources.DeepCopy(),
		Pods:               make([]*client.ClusterPod, len(oldNode.Pods)-1),
	}

	var removedPod *client.ClusterPod
	destIndex := 0
	for _, currPod := range oldNode.Pods {
		if currPod.Name != podToRemove.Name || currPod.Namespace != podToRemove.Namespace {
			updatedNode.Pods[destIndex] = currPod
			destIndex++
		} else {
			removedPod = currPod
		}
	}

	updatedNode.AvailableResources.Add(removedPod.TotalResources)
	return updatedNode
}

// Updates the specified pod on a copy of the clusterNode and returns that updated copy.
func (knc *KubernetesNodesCache) updatePodOnNode(oldNode *client.ClusterNode, podToUpdate *core.Pod) *client.ClusterNode {
	updatedPod := client.NewClusterPod(podToUpdate)
	newPods := make([]*client.ClusterPod, len(oldNode.Pods), len(oldNode.Pods)+1)

	podUpdated := false
	for i, currPod := range oldNode.Pods {
		if currPod.Name != podToUpdate.Name || currPod.Namespace != podToUpdate.Namespace {
			newPods[i] = currPod
		} else {
			newPods[i] = updatedPod
			podUpdated = true
		}
	}

	// Ensure that we also handle the strange case, where client-go calls the update callback also for additions.
	if !podUpdated {
		newPods = append(newPods, updatedPod)
	}

	return client.NewClusterNodeWithPods(oldNode.Node, newPods)
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
