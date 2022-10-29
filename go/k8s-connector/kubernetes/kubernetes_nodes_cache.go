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

type nodesCacheUpdate struct {
	updateType updateType
	node       *core.Node
}

// NodesCache implementation for Kubernetes.
type KubernetesNodesCache struct {
	ctx context.Context

	store          collections.ConcurrentObjectStore[*core.Node]
	updateInterval time.Duration
	updateQueue    chan *nodesCacheUpdate

	clusterClient KubernetesClusterClient
	informer      cache.SharedIndexInformer
}

func NewKubernetesNodesCache(
	clusterClient KubernetesClusterClient,
	updateInterval time.Duration,
	queueSize int,
) *KubernetesNodesCache {
	knc := &KubernetesNodesCache{
		store:          collections.NewConcurrentObjectStoreImpl[*core.Node](),
		clusterClient:  clusterClient,
		updateInterval: updateInterval,
		updateQueue:    make(chan *nodesCacheUpdate, queueSize),
	}

	return knc
}

func (knc *KubernetesNodesCache) Nodes() collections.ConcurrentObjectStore[*core.Node] {
	return knc.store
}

func (knc *KubernetesNodesCache) StartWatch(ctx context.Context) error {
	if knc.ctx != nil {
		return fmt.Errorf("watch has already been started")
	}
	knc.ctx = ctx

	knc.informer = knc.setUpInformer(knc.clusterClient)
	go knc.runUpdateLoop()
	return knc.startInformerAndDoInitialSync(knc.informer)
}

func (knc *KubernetesNodesCache) setUpInformer(clusterClient KubernetesClusterClient) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(clusterClient.ClientSet(), 0)
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
	var storeWriter collections.ConcurrentObjectStoreWriter[*core.Node]

	for {
		select {
		case update := <-knc.updateQueue:
			if storeWriter == nil {
				storeWriter = knc.store.WriteLock()
				defer storeWriter.Unlock()
			}
			knc.applyUpdate(update, storeWriter)
		default:
			return
		}
	}
}

func (knc *KubernetesNodesCache) applyUpdate(update *nodesCacheUpdate, storeWriter collections.ConcurrentObjectStoreWriter[*core.Node]) {
	if update.updateType == removal {
		storeWriter.Remove(update.node.Name)
	} else {
		storeWriter.Set(update.node.Name, update.node)
	}
}

func coerceToNodeOrPanic(obj interface{}) *core.Node {
	node, ok := obj.(*core.Node)
	if !ok {
		panic("NodeInformer received a non Node object")
	}
	return node
}
