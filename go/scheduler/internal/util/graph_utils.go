package util

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

// GetServiceGraphNode gets the node from the ServiceGraph, which corresponds to the specified pod.
func GetServiceGraphNode(svcGraph *servicegraph.ServiceGraph, pod *v1.Pod) (*servicegraph.MicroserviceNode, error) {
	microserviceLabel, err := GetPodInstanceLabel(pod)
	if err != nil {
		return nil, err
	}

	microserviceNode := svcGraph.NodeByLabel(microserviceLabel)
	if microserviceNode == nil {
		return nil, fmt.Errorf("No microservice node matching the pod's instance label found in the ServiceGraph")
	}

	return microserviceNode, nil
}
