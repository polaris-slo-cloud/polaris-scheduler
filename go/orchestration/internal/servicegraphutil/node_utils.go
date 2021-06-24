package servicegraphutil

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

// CreatePodSpec creates a PodSpec from the specified node.
func CreatePodTemplate(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*core.PodTemplateSpec, error) {
	podTemplate := core.PodTemplateSpec{
		ObjectMeta: meta.ObjectMeta{},
		Spec:       core.PodSpec{},
	}

	return &podTemplate, nil
}

// CreateDeployment creates a new Deployment from the specified node.
func CreateDeployment(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*apps.Deployment, error) {
	deployment := apps.Deployment{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       apps.DeploymentSpec{},
	}

	return UpdateDeployment(&deployment, node, graph)
}

// UpdateDeployment updates an existing Deployment, based on the specified node.
func UpdateDeployment(deployment *apps.Deployment, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*apps.Deployment, error) {
	replicas := getInitialReplicas(node)

	updateNodeObjectMeta(&deployment.ObjectMeta, node, graph)
	updatePodTemplate(&deployment.Spec.Template, node, graph)
	deployment.Spec.Selector = createLabelSelector(node, graph)
	deployment.Spec.Replicas = &replicas

	return deployment, nil
}

// CreateStatefulSet creates a StatefulSet from the specified node.
func CreateStatefulSet(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*apps.StatefulSet, error) {
	statefulSet := apps.StatefulSet{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       apps.StatefulSetSpec{},
	}

	return UpdateStatefulSet(&statefulSet, node, graph)
}

// UpdateStatefulSet updates an existing StatefulSet, based on the specified node.
func UpdateStatefulSet(statefulSet *apps.StatefulSet, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*apps.StatefulSet, error) {
	replicas := getInitialReplicas(node)

	updateNodeObjectMeta(&statefulSet.ObjectMeta, node, graph)
	updatePodTemplate(&statefulSet.Spec.Template, node, graph)
	statefulSet.Spec.Selector = createLabelSelector(node, graph)
	statefulSet.Spec.Replicas = &replicas

	return statefulSet, nil
}

func updatePodTemplate(podTemplate *core.PodTemplateSpec, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) {
	podTemplate.Spec.SchedulerName = kubeutil.RainbowSchedulerName
	podTemplate.ObjectMeta.Labels = getPodLabels(node, graph)
	podTemplate.Spec.InitContainers = node.InitContainers
	podTemplate.Spec.Containers = node.Containers
	podTemplate.Spec.Volumes = node.Volumes

	if serviceAccountName := getServiceAccountName(node, graph); serviceAccountName != nil {
		podTemplate.Spec.ServiceAccountName = *serviceAccountName
	} else {
		podTemplate.Spec.ServiceAccountName = ""
	}
}

func createLabelSelector(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *meta.LabelSelector {
	return &meta.LabelSelector{
		MatchLabels: getPodLabels(node, graph),
	}
}

func getInitialReplicas(node *fogappsCRDs.ServiceGraphNode) int32 {
	if node.Replicas.InitialCount != nil {
		return *node.Replicas.InitialCount
	}
	return node.Replicas.Min
}

func getServiceAccountName(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *string {
	if node.ServiceAccountName != nil {
		return node.ServiceAccountName
	}
	return graph.Spec.ServiceAccountName
}
