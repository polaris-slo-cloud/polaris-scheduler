package servicegraphutil

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/internal/util"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

// createNodeObjectMeta creates an ObjectMeta for resources that are created from a ServiceGraphNode.
//
// The returned object needs to be passed to updateNodeObjectMeta() as well to set the updateable fields.
func createNodeObjectMeta(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *meta.ObjectMeta {
	return &meta.ObjectMeta{
		Name:        node.Name,
		Namespace:   graph.Namespace,
		Labels:      getPodLabels(node, graph),
		Annotations: make(map[string]string),
	}
}

// updateNodeObjectMeta updates an existing ObjectMeta for ServiceGraphNode derived resources.
func updateNodeObjectMeta(objectMeta *meta.ObjectMeta, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) {
	objectMeta.Labels = getPodLabels(node, graph)
}

// getPodLabels gets the labels for a pod generated from a ServiceGraphNode.
func getPodLabels(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) map[string]string {
	labels := util.DeepCopyStringMap(node.PodLabels)
	labels[kubeutil.LabelRainbowGeneratedPod] = fmt.Sprintf("%s.%s.generated", graph.Name, node.Name)
	labels[kubeutil.LabelRefServiceGraph] = graph.Name
	labels[kubeutil.LabelRefServiceGraphNode] = node.Name
	return labels
}
