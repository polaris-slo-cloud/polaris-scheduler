package kubernetes

import (
	"fmt"

	"github.com/go-logr/logr"

	core "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/config"
)

var (
	_ client.ClusterClient = (*KubernetesClusterClient)(nil)
)

// ClusterClient implementation for Kubernetes.
type KubernetesClusterClient struct {
	clusterName   string
	k8sClientSet  *clientset.Clientset
	eventRecorder record.EventRecorder
}

// Creates a new KubernetesClusterClient using the specified kubeconfig.
func NewKubernetesClusterClient(
	clusterName string,
	kubeconfig *rest.Config,
	schedConfig *config.SchedulerConfig,
	logger *logr.Logger,
) (*KubernetesClusterClient, error) {
	k8sClientSet, err := clientset.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	eventSink := coreclient.EventSinkImpl{Interface: k8sClientSet.CoreV1().Events("")}
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(
		func(format string, args ...interface{}) {
			msg := fmt.Sprintf(format, args)
			logger.Info(msg)
		},
	)
	eventBroadcaster.StartRecordingToSink(&eventSink)
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: schedConfig.SchedulerName})

	clusterClient := KubernetesClusterClient{
		clusterName:   clusterName,
		k8sClientSet:  k8sClientSet,
		eventRecorder: eventRecorder,
	}

	return &clusterClient, nil
}

func (c *KubernetesClusterClient) ClusterName() string {
	return c.clusterName
}

func (c *KubernetesClusterClient) ClientSet() clientset.Interface {
	return c.k8sClientSet
}

func (c *KubernetesClusterClient) EventRecorder() record.EventRecorder {
	return c.eventRecorder
}
