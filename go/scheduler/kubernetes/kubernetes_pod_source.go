package kubernetes

import (
	"fmt"

	core "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/pipeline"
)

var (
	_ pipeline.PodSource = (*KubernetesPodSource)(nil)
)

type KubernetesPodSource struct {
	clientsMgr      client.ClusterClientsManager
	incomingPods    chan *core.Pod
	schedConfig     *config.SchedulerConfig
	stopChan        chan struct{}
	sharedInformers map[string]cache.SharedIndexInformer
}

// Creates a new KubernetesPodSource for all clusters in the specified ClusterClientsManager.
func NewKubernetesPodSource(clusterClientsMgr client.ClusterClientsManager, schedConfig *config.SchedulerConfig) *KubernetesPodSource {
	kps := KubernetesPodSource{
		clientsMgr:      clusterClientsMgr,
		incomingPods:    make(chan *core.Pod, schedConfig.IncomingPodsBufferSize),
		schedConfig:     schedConfig,
		sharedInformers: make(map[string]cache.SharedIndexInformer, clusterClientsMgr.ClustersCount()),
	}
	return &kps
}

// Creates the channel available through IncomingPods() and starts watching for pods.
func (kps *KubernetesPodSource) StartWatching() error {
	if kps.stopChan != nil {
		return fmt.Errorf("this KubernetesPodSource is already watching")
	}

	kps.stopChan = make(chan struct{}, 1)
	kps.setUpInformers()

	// Start all informers
	for _, informer := range kps.sharedInformers {
		go informer.Run(kps.stopChan)
	}

	// Wait for each informer to be synced.
	for _, informer := range kps.sharedInformers {
		if !cache.WaitForCacheSync(kps.stopChan, informer.HasSynced) {
			err := fmt.Errorf("timed out waiting for caches to sync")
			runtime.HandleError(err)
			return err
		}
	}

	return nil
}

// Stops watching for pods and closes the channel available through IncomingPods().
func (kps *KubernetesPodSource) StopWatching() error {
	close(kps.stopChan)
	return nil
}

func (kps *KubernetesPodSource) IncomingPods() chan *core.Pod {
	return kps.incomingPods
}

func (kps *KubernetesPodSource) setUpInformers() {
	kps.clientsMgr.ForEach(func(clusterName string, client client.ClusterClient) {
		kps.sharedInformers[clusterName] = kps.setUpInformer(client)
	})
}

func (kps *KubernetesPodSource) setUpInformer(clusterClient client.ClusterClient) cache.SharedIndexInformer {
	factory := informers.NewSharedInformerFactory(clusterClient.ClientSet(), 0)
	podInformer := factory.Core().V1().Pods()
	sharedInformer := podInformer.Informer()

	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*core.Pod)
			if !ok {
				panic("PodInformer received a non Pod object")
			}
			kps.onAdd(pod)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod, ok := oldObj.(*core.Pod)
			if !ok {
				panic("PodInformer received a non Pod object")
			}
			newPod, ok := newObj.(*core.Pod)
			if !ok {
				panic("PodInformer received a non Pod object")
			}
			kps.onUpdate(oldPod, newPod)
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*core.Pod)
			if !ok {
				panic("PodInformer received a non Pod object")
			}
			kps.onDelete(pod)
		},
	})

	return sharedInformer
}

func (kps *KubernetesPodSource) onAdd(pod *core.Pod) {
	kps.publishPodIfUnscheduled(pod)
}

func (kps *KubernetesPodSource) onUpdate(oldPod, newPod *core.Pod) {
	kps.publishPodIfUnscheduled(newPod)
}

func (kps *KubernetesPodSource) onDelete(pod *core.Pod) {
	// We ignore deletions for now.
}

func (kps *KubernetesPodSource) publishPodIfUnscheduled(pod *core.Pod) {
	if pod.Spec.NodeName == "" && pod.Spec.SchedulerName == kps.schedConfig.SchedulerName {
		kps.incomingPods <- pod
	}
}
