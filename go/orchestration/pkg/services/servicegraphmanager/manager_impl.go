package servicegraphmanager

import (
	"gonum.org/v1/gonum/graph"
	v1 "k8s.io/api/core/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/util"
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

func (me *serviceGraphManagerImpl) buildServiceGraph(pod *v1.Pod) (*servicegraph.ServiceGraph, error) {
	appName, err := util.GetAppName(pod)
	if err != nil {
		return nil, err
	}
	namespace := util.GetNamespace(&pod.ObjectMeta)
	svcGraph := servicegraph.NewServiceGraph(namespace, appName)
	svcGraph.SetMaxDelayMs(800)

	mqNode := svcGraph.AddNewNode(
		"message-queue-0",
		&servicegraph.MicroserviceNodeInfo{
			MicroserviceType:           util.MicroserviceTypeMessageQueue,
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

	edges := []graph.WeightedEdge{
		svcGraph.NewWeightedEdge(mqNode, taxiCloudNode, 1),
		svcGraph.NewWeightedEdge(mqNode, taxiIoTNode, 1),
		svcGraph.NewWeightedEdge(mqNode, taxiEdgeNodes[0], 1),
		svcGraph.NewWeightedEdge(mqNode, taxiEdgeNodes[1], 1),
	}
	for _, edge := range edges {
		svcGraph.SetWeightedEdge(edge)
	}

	return svcGraph, nil
}

func (me *serviceGraphManagerImpl) updateMaxDelay(svcGraph *servicegraph.ServiceGraph, pod *v1.Pod) {
	label, err := util.GetPodInstanceLabel(pod)
	if err != nil {
		return
	}

	node := svcGraph.NodeByLabel(label)
	if node == nil {
		return
	}

	maxDelayMs := util.GetPodMaxDelay(pod)
	svcGraph.Mutex.Lock()
	node.MicroserviceNodeInfo().MaxLatencyToMessageQueueMs = maxDelayMs
	svcGraph.Mutex.Unlock()
}
