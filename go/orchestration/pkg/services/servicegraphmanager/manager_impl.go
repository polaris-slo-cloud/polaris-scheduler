package servicegraphmanager

import (
	"fmt"
	"math"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/labeledgraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

var (
	_serviceGraphManagerImpl *serviceGraphManagerImpl

	_ ServiceGraphManager = _serviceGraphManagerImpl
)

type serviceGraphManagerImpl struct {
	mockServiceGraph *servicegraph.ServiceGraph
}

func newServiceGraphManagerImpl() *serviceGraphManagerImpl {
	return &serviceGraphManagerImpl{}
}

func (me *serviceGraphManagerImpl) ServiceGraph(pod *v1.Pod) (*servicegraph.ServiceGraph, error) {
	if me.mockServiceGraph == nil {
		if svcGraph, err := me.buildServiceGraph(pod); err == nil {
			me.mockServiceGraph = svcGraph
		} else {
			return nil, err
		}
	}
	me.updateMaxDelay(me.mockServiceGraph, pod)
	return me.mockServiceGraph, nil
}

// ToDo: Lots of mocked stuff here - remove that.

func (me *serviceGraphManagerImpl) buildServiceGraph(pod *v1.Pod) (*servicegraph.ServiceGraph, error) {
	appName, err := getAppName(pod)
	if err != nil {
		return nil, err
	}
	namespace := kubeutil.GetNamespace(&pod.ObjectMeta)
	svcGraph := servicegraph.NewServiceGraph(namespace, appName)
	svcGraph.SetMaxDelayMs(800)

	mqNode := svcGraph.AddNewNode(
		"message-queue-0",
		&servicegraph.MicroserviceNodeInfo{
			MicroserviceType:           microserviceTypeMessageQueue,
			MaxLatencyToMessageQueueMs: 0,
		},
	)
	svcGraph.SetMessageQueueNode(mqNode)

	taxiCloudNode := svcGraph.AddNewNode(
		"taxi-cloud-0",
		&servicegraph.MicroserviceNodeInfo{
			MicroserviceType:           "taxi-cloud",
			MaxLatencyToMessageQueueMs: 100,
		},
	)

	taxiIoTNode := svcGraph.AddNewNode(
		"taxi-iot-0",
		&servicegraph.MicroserviceNodeInfo{
			MicroserviceType:           "taxi-iot",
			MaxLatencyToMessageQueueMs: 40,
		},
	)

	taxiEdgeNodes := []*servicegraph.MicroserviceNode{
		svcGraph.AddNewNode(
			"taxi-edge-bronx",
			&servicegraph.MicroserviceNodeInfo{
				MicroserviceType:           "taxi-edge",
				MaxLatencyToMessageQueueMs: 50,
			},
		),
		svcGraph.AddNewNode(
			"taxi-edge-brooklyn",
			&servicegraph.MicroserviceNodeInfo{
				MicroserviceType:           "taxi-edge",
				MaxLatencyToMessageQueueMs: 50,
			},
		),
	}

	edges := []labeledgraph.WeightedEdge{
		svcGraph.NewWeightedEdge(mqNode, taxiCloudNode, labeledgraph.NewComplexEdgeWeightFromFloat(1)),
		svcGraph.NewWeightedEdge(mqNode, taxiIoTNode, labeledgraph.NewComplexEdgeWeightFromFloat(1)),
		svcGraph.NewWeightedEdge(mqNode, taxiEdgeNodes[0], labeledgraph.NewComplexEdgeWeightFromFloat(1)),
		svcGraph.NewWeightedEdge(mqNode, taxiEdgeNodes[1], labeledgraph.NewComplexEdgeWeightFromFloat(1)),
	}
	for _, edge := range edges {
		svcGraph.SetWeightedEdge(edge)
	}

	return svcGraph, nil
}

func (me *serviceGraphManagerImpl) updateMaxDelay(svcGraph *servicegraph.ServiceGraph, pod *v1.Pod) {
	label, err := getPodInstanceLabel(pod)
	if err != nil {
		return
	}

	node := svcGraph.NodeByLabel(label)
	if node == nil {
		return
	}

	maxDelayMs := getPodMaxDelay(pod)
	svcGraph.Mutex.Lock()
	node.MicroserviceNodeInfo().MaxLatencyToMessageQueueMs = maxDelayMs
	svcGraph.Mutex.Unlock()
}

const (
	// microserviceTypeMessageQueue is the string constant used to identify a message queue pod.
	microserviceTypeMessageQueue = "message-queue"

	microserviceTypeLabel = "app.kubernetes.io/component"
	appNameLabel          = "app.kubernetes.io/name"
	instanceNameLabel     = "app.kubernetes.io/instance"
	maxDelayMsLabel       = "rainbow-h2020.eu/max-delay-ms"
)

// GetPodMicroserviceType returns the type of microservice that the pod is supposed to host.
func getPodMicroserviceType(pod *v1.Pod) (string, bool) {
	return kubeutil.GetLabel(&pod.ObjectMeta, microserviceTypeLabel)
}

// isPodMessageQueue returns true if the specified pod is supposed to host a message queue.
func isPodMessageQueue(pod *v1.Pod) bool {
	msType, exists := getPodMicroserviceType(pod)
	return exists && msType == microserviceTypeMessageQueue
}

// GetAppName returns the name of the app that the pod belongs to.
func getAppName(pod *v1.Pod) (string, error) {
	appName, ok := kubeutil.GetLabel(&pod.ObjectMeta, appNameLabel)
	if ok {
		return appName, nil
	}
	return appName, fmt.Errorf("The pod has no %s label", appNameLabel)
}

// GetPodMaxDelay gets the max delay in milliseconds that has been configured for the pod.
// If no max delay is defined for the Pod, a default value (MaxInt64) is returned.
func getPodMaxDelay(pod *v1.Pod) int64 {
	delayMsStr, ok := kubeutil.GetLabel(&pod.ObjectMeta, maxDelayMsLabel)
	if ok {
		maxDelay, err := strconv.ParseInt(delayMsStr, 10, 64)
		if err == nil {
			return maxDelay
		}
	}
	return math.MaxInt64
}

// GetPodInstanceLabel gets the instance label from the pod.
// This is used to identify the pod's not in the ServiceGraph.
func getPodInstanceLabel(pod *v1.Pod) (string, error) {
	instanceLabel, ok := kubeutil.GetLabel(&pod.ObjectMeta, instanceNameLabel)
	if ok {
		return instanceLabel, nil
	}
	return instanceLabel, fmt.Errorf("The pod has no %s label", instanceNameLabel)
}
