package util

import (
	"fmt"
	"math"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

const (
	// MicroserviceTypeMessageQueue is the string constant used to identify a message queue pod.
	MicroserviceTypeMessageQueue = "message-queue"

	microserviceTypeLabel = "app.kubernetes.io/component"
	appNameLabel          = "app.kubernetes.io/name"
	instanceNameLabel     = "app.kubernetes.io/instance"
	maxDelayMsLabel       = "rainbow-h2020.eu/max-delay-ms"
)

// GetPodMicroserviceType returns the type of microservice that the pod is supposed to host.
func GetPodMicroserviceType(pod *v1.Pod) (string, bool) {
	return kubeutil.GetLabel(&pod.ObjectMeta, microserviceTypeLabel)
}

// IsPodMessageQueue returns true if the specified pod is supposed to host a message queue.
func IsPodMessageQueue(pod *v1.Pod) bool {
	msType, exists := GetPodMicroserviceType(pod)
	return exists && msType == MicroserviceTypeMessageQueue
}

// GetAppName returns the name of the app that the pod belongs to.
func GetAppName(pod *v1.Pod) (string, error) {
	appName, ok := kubeutil.GetLabel(&pod.ObjectMeta, appNameLabel)
	if ok {
		return appName, nil
	}
	return appName, fmt.Errorf("The pod has no %s label", appNameLabel)
}

// GetPodMaxDelay gets the max delay in milliseconds that has been configured for the pod.
// If no max delay is defined for the Pod, a default value (MaxInt64) is returned.
func GetPodMaxDelay(pod *v1.Pod) int64 {
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
func GetPodInstanceLabel(pod *v1.Pod) (string, error) {
	instanceLabel, ok := kubeutil.GetLabel(&pod.ObjectMeta, instanceNameLabel)
	if ok {
		return instanceLabel, nil
	}
	return instanceLabel, fmt.Errorf("The pod has no %s label", instanceNameLabel)
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
