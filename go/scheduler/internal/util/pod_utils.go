package util

import (
	"fmt"

	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

// GetPodServiceGraphNodeName gets the name of the ServiceGraphNode from the pod.
func GetPodServiceGraphNodeName(pod *core.Pod) (string, bool) {
	return kubeutil.GetLabel(pod, kubeutil.LabelRefServiceGraphNode)
}

// CalcTotalRequiredResources calculated the total resources required by the pod.
func CalcTotalRequiredResources(pod *v1.Pod) (*framework.Resource, error) {
	required := &framework.Resource{}
	for _, container := range pod.Spec.Containers {
		if container.Resources.Limits == nil {
			return nil, fmt.Errorf("Cannot schedule pod %s, because container %s did not specify any resource limits", pod.Name, container.Name)
		}
		required.Add(container.Resources.Limits)
	}
	return required, nil
}
