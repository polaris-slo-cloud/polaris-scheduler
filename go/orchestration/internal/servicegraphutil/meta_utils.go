package servicegraphutil

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogapps "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/internal/util"
)

const (
	RainbowGeneratedPodLabelName = "rainbow-generated-pod-label"
)

// createNodeObjectMeta creates an ObjectMeta for resources that are created from a ServiceGraphNode.
func createNodeObjectMeta(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) *meta.ObjectMeta {
	return &meta.ObjectMeta{
		Name:        node.Name,
		Namespace:   graph.Namespace,
		Labels:      getPodLabels(node, graph),
		Annotations: make(map[string]string),
	}
}

// getPodLabels gets the labels for a pod generated from a ServiceGraphNode.
func getPodLabels(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) map[string]string {
	labels := util.DeepCopyStringMap(node.PodLabels)
	if len(labels) == 0 {
		labels[RainbowGeneratedPodLabelName] = fmt.Sprintf("%s.%s.generated", graph.Name, node.Name)
	}
	return labels
}
