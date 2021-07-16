package servicegraphutil

import (
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
)

const (
	kubernetesCpuArchLabel = "kubernetes.io/arch"
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
	updateNodeObjectMeta(&deployment.ObjectMeta, node, graph)
	updatePodTemplate(&deployment.Spec.Template, node, graph)
	deployment.Spec.Selector = createLabelSelector(node, graph)

	initialReplicas := GetInitialReplicas(node)
	initialReplicasFromStatus := getInitialReplicasFromStatus(node, graph)
	if initialReplicasFromStatus != initialReplicas {
		deployment.Spec.Replicas = &initialReplicas
	}

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
	updateNodeObjectMeta(&statefulSet.ObjectMeta, node, graph)
	updatePodTemplate(&statefulSet.Spec.Template, node, graph)
	statefulSet.Spec.Selector = createLabelSelector(node, graph)

	initialReplicas := GetInitialReplicas(node)
	initialReplicasFromStatus := getInitialReplicasFromStatus(node, graph)
	if initialReplicasFromStatus != initialReplicas {
		statefulSet.Spec.Replicas = &initialReplicas
	}

	return statefulSet, nil
}

func updatePodTemplate(podTemplate *core.PodTemplateSpec, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) {
	podTemplate.Spec.SchedulerName = kubeutil.RainbowSchedulerName
	podTemplate.ObjectMeta.Labels = getPodLabels(node, graph)
	podTemplate.Spec.InitContainers = node.InitContainers
	podTemplate.Spec.Containers = node.Containers
	podTemplate.Spec.Volumes = node.Volumes
	podTemplate.Spec.Affinity = node.Affinity

	if node.ImagePullSecrets != nil && len(node.ImagePullSecrets) > 0 {
		podTemplate.Spec.ImagePullSecrets = node.ImagePullSecrets
	} else {
		podTemplate.Spec.ImagePullSecrets = nil
	}

	if node.NodeHardware != nil {
		addNodeHardwareRequirements(podTemplate, node.NodeHardware)
	}
	// ToDo: Add else branch to remove hardware requirements from the pod, if they were removed from the ServiceGraphNode.

	if serviceAccountName := getServiceAccountName(node, graph); serviceAccountName != nil {
		podTemplate.Spec.ServiceAccountName = *serviceAccountName
	} else {
		podTemplate.Spec.ServiceAccountName = ""
	}

	if node.ExposedPorts != nil {
		podTemplate.Spec.HostNetwork = node.ExposedPorts.HostNetwork
	} else {
		// Set to the default value.
		podTemplate.Spec.HostNetwork = false
	}

	if graph.Spec.DNSConfig != nil {
		podTemplate.Spec.DNSPolicy = graph.Spec.DNSConfig.DNSPolicy
		podTemplate.Spec.DNSConfig = &graph.Spec.DNSConfig.PodDNSConfig
	} else {
		podTemplate.Spec.DNSPolicy = core.DNSClusterFirst
		podTemplate.Spec.DNSConfig = nil
	}
}

func createLabelSelector(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *meta.LabelSelector {
	return &meta.LabelSelector{
		MatchLabels: getPodLabels(node, graph),
	}
}

// GetInitialReplicas returns the initial number of replicas configured for the node
// or, if they are not set, the minimum number of replicas.
func GetInitialReplicas(node *fogappsCRDs.ServiceGraphNode) int32 {
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

// addNodeHardwareRequirements adds/configures the affinity of the podTemplate to ensure that only nodes
// that match the nodeHardwareReq are eligible.
func addNodeHardwareRequirements(podTemplate *core.PodTemplateSpec, nodeHardwareReq *fogappsCRDs.NodeHardware) {
	if nodeHardwareReq.NodeType == nil && nodeHardwareReq.CpuInfo == nil && nodeHardwareReq.GpuInfo == nil {
		return
	}

	nodeSelector := ensureAffinityNodeSelectorExists(podTemplate)

	if nodeHardwareReq.CpuInfo != nil {
		addCpuSelectionTerms(nodeSelector, nodeHardwareReq.CpuInfo)
	}

	if len(nodeSelector.NodeSelectorTerms) == 0 {
		nodeSelector.NodeSelectorTerms = nil
	}

	// ToDo: Add support for other NodeHardware fields!
}

// ensureAffinityNodeSelectorExists creates the required during scheduling NodeSelector for the pod's node affinity,
// if it does not exist, and returns the NodeSelector.
func ensureAffinityNodeSelectorExists(podTemplate *core.PodTemplateSpec) *core.NodeSelector {
	if podTemplate.Spec.Affinity == nil {
		podTemplate.Spec.Affinity = &core.Affinity{}
	}
	if podTemplate.Spec.Affinity.NodeAffinity == nil {
		podTemplate.Spec.Affinity.NodeAffinity = &core.NodeAffinity{}
	}
	if podTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		podTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &core.NodeSelector{}
	}
	nodeSelector := podTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
	if nodeSelector.NodeSelectorTerms == nil {
		nodeSelector.NodeSelectorTerms = make([]core.NodeSelectorTerm, 0)
	}
	return nodeSelector
}

func addCpuSelectionTerms(nodeSelector *core.NodeSelector, cpuInfo *fogappsCRDs.CpuInfo) {
	architecturesCount := len(cpuInfo.Architectures)
	if architecturesCount == 0 {
		return
	}

	architectures := make([]string, architecturesCount)
	for i, cpuArch := range cpuInfo.Architectures {
		architectures[i] = string(cpuArch)
	}

	cpuArchReq := core.NodeSelectorRequirement{
		Key:      kubernetesCpuArchLabel,
		Operator: core.NodeSelectorOpIn,
		Values:   architectures,
	}

	if len(nodeSelector.NodeSelectorTerms) == 0 {
		nodeSelector.NodeSelectorTerms = append(nodeSelector.NodeSelectorTerms, core.NodeSelectorTerm{})
	}

	// Add cpuArchReq to all node selector terms, because only one needs to be satisfied by a node to be eligible.
	for i := range nodeSelector.NodeSelectorTerms {
		selectorTerm := &nodeSelector.NodeSelectorTerms[i]
		if selectorTerm.MatchExpressions == nil {
			selectorTerm.MatchExpressions = make([]core.NodeSelectorRequirement, 0, 1)
		}
		selectorTerm.MatchExpressions = append(selectorTerm.MatchExpressions, cpuArchReq)
	}
}

// Returns the initial replica count for the specified node from the Status of the ServiceGraph,
// or -1 if the node does not exist in the Status.
func getInitialReplicasFromStatus(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) int32 {
	if graph.Status.NodeStates == nil {
		return -1
	}
	if nodeState, ok := graph.Status.NodeStates[node.Name]; ok {
		if nodeState.InitialReplicas > 0 {
			return nodeState.InitialReplicas
		}
	}
	return -1
}
