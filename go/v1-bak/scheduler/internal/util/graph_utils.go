package util

import (
	"fmt"

	core "k8s.io/api/core/v1"

	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/model/graph/servicegraph"
)

// GetServiceGraphNode gets the node from the ServiceGraph, which corresponds to the specified pod.
func GetServiceGraphNode(svcGraph servicegraph.ServiceGraph, pod *core.Pod) (servicegraph.Node, error) {
	svcGraphNodeName, ok := GetPodServiceGraphNodeName(pod)
	if !ok {
		return nil, fmt.Errorf("The pod does not have a label %s", kubeutil.LabelRefServiceGraphNode)
	}

	svcGraphNode := svcGraph.NodeByLabel(svcGraphNodeName)
	if svcGraphNode == nil {
		return nil, fmt.Errorf("No ServiceGraph Node found that matches the pod's label %s", kubeutil.LabelRefServiceGraphNode)
	}

	return svcGraphNode, nil
}

// GetServiceGraphCRDNode gets the CRD node from the ServiceGraph CRD instance, which corresponds to the specified pod.
func GetServiceGraphCRDNode(svcGraphCRD *fogappsCRDs.ServiceGraph, pod *core.Pod) (*fogappsCRDs.ServiceGraphNode, error) {
	svcGraphNodeName, ok := GetPodServiceGraphNodeName(pod)
	if !ok {
		return nil, fmt.Errorf("The pod does not have a label %s", kubeutil.LabelRefServiceGraphNode)
	}

	for i := range svcGraphCRD.Spec.Nodes {
		node := &svcGraphCRD.Spec.Nodes[i]
		if node.Name == svcGraphNodeName {
			return node, nil
		}
	}

	return nil, fmt.Errorf("No ServiceGraphNode found that matches the pod's label %s", kubeutil.LabelRefServiceGraphNode)
}
