package kubernetes

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/client"
)

var (
	_ client.ClusterClient = (*KubernetesClusterClientImpl)(nil)
)

// ClusterClient implementation for Kubernetes.
type KubernetesClusterClientImpl struct {
	clusterName   string
	k8sClientSet  *clientset.Clientset
	eventRecorder record.EventRecorder
	logger        *logr.Logger
}

// Creates a new KubernetesClusterClientImpl using the specified kubeconfig.
//
// - clusterName is the name of the cluster to connect to
// - kubeconfig is the respective kubeconfig
// - parentComponentName is the name of the component that creates this client (this is used as the source name in the event recorder)
// - logger the Logger that should be used for logging
func NewKubernetesClusterClientImpl(
	clusterName string,
	kubeconfig *rest.Config,
	parentComponentName string,
	logger *logr.Logger,
) (*KubernetesClusterClientImpl, error) {
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
	eventRecorder := eventBroadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: parentComponentName})

	clusterClient := KubernetesClusterClientImpl{
		clusterName:   clusterName,
		k8sClientSet:  k8sClientSet,
		eventRecorder: eventRecorder,
		logger:        logger,
	}

	return &clusterClient, nil
}

func (c *KubernetesClusterClientImpl) ClusterName() string {
	return c.clusterName
}

func (c *KubernetesClusterClientImpl) ClientSet() clientset.Interface {
	return c.k8sClientSet
}

func (c *KubernetesClusterClientImpl) EventRecorder() record.EventRecorder {
	return c.eventRecorder
}

func (c *KubernetesClusterClientImpl) CommitSchedulingDecision(ctx context.Context, pod *core.Pod, binding *core.Binding) error {
	err := c.k8sClientSet.CoreV1().Pods(pod.Namespace).Bind(ctx, binding, meta.CreateOptions{})
	if err != nil {
		c.logger.Error(err, "could not bind Pod", "pod", pod, "binding", binding)
		c.eventRecorder.Eventf(pod, "Error", "Could not bind Pod", "Could not bind pod to node %s", &binding.Target.Name)
		return err
	}

	fullyQualifiedPodName := pod.Namespace + "." + pod.Name
	c.logger.Info("PodScheduled", "pod", fullyQualifiedPodName, "cluster", c.clusterName, "node", binding.Target.Name)
	return nil
}
