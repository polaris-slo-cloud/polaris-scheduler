package servicegraphutil

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogapps "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

// CreatePodSpec creates a PodSpec from the specified node.
func CreatePodTemplate(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (*core.PodTemplateSpec, error) {
	podTemplate := core.PodTemplateSpec{
		ObjectMeta: meta.ObjectMeta{},
		Spec:       core.PodSpec{},
	}

	return &podTemplate, nil
}

// CreateDeployment creates a new Deployment from the specified node.
func CreateDeployment(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (*apps.Deployment, error) {
	deployment := apps.Deployment{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       apps.DeploymentSpec{},
	}

	return UpdateDeployment(&deployment, node, graph)
}

// UpdateDeployment updates an existing Deployment, based on the specified node.
func UpdateDeployment(deployment *apps.Deployment, node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (*apps.Deployment, error) {
	replicas := getInitialReplicas(node)

	updateNodeObjectMeta(&deployment.ObjectMeta, node, graph)
	updatePodTemplate(&deployment.Spec.Template, node, graph)
	deployment.Spec.Selector = createLabelSelector(node, graph)
	deployment.Spec.Replicas = &replicas

	return deployment, nil
}

// CreateStatefulSet creates a StatefulSet from the specified node.
func CreateStatefulSet(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (*apps.StatefulSet, error) {
	statefulSet := apps.StatefulSet{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       apps.StatefulSetSpec{},
	}

	return UpdateStatefulSet(&statefulSet, node, graph)
}

// UpdateStatefulSet updates an existing StatefulSet, based on the specified node.
func UpdateStatefulSet(statefulSet *apps.StatefulSet, node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (*apps.StatefulSet, error) {
	replicas := getInitialReplicas(node)

	updateNodeObjectMeta(&statefulSet.ObjectMeta, node, graph)
	updatePodTemplate(&statefulSet.Spec.Template, node, graph)
	statefulSet.Spec.Selector = createLabelSelector(node, graph)
	statefulSet.Spec.Replicas = &replicas

	return statefulSet, nil
}

func updatePodTemplate(podTemplate *core.PodTemplateSpec, node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) {
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

func createLabelSelector(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) *meta.LabelSelector {
	return &meta.LabelSelector{
		MatchLabels: getPodLabels(node, graph),
	}
}

func getInitialReplicas(node *fogapps.ServiceGraphNode) int32 {
	if node.Replicas.InitialCount != nil {
		return *node.Replicas.InitialCount
	}
	return node.Replicas.Min
}

func getServiceAccountName(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) *string {
	if node.ServiceAccountName != nil {
		return node.ServiceAccountName
	}
	return graph.Spec.ServiceAccountName
}
