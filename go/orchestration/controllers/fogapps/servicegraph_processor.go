package fogapps

import (
	"fmt"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	svcGraphUtil "k8s.rainbow-h2020.eu/rainbow/orchestration/internal/servicegraphutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/controllerutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/kubeutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/slo"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceGraphChildObjectMaps struct {
	Deployments  map[string]*apps.Deployment
	StatefulSets map[string]*apps.StatefulSet
	Services     map[string]*core.Service
	Ingresses    map[string]*networking.Ingress
	SloMappings  map[string]*slo.UnstructuredSloMapping
}

type serviceGraphProcessor struct {
	svcGraph *fogappsCRDs.ServiceGraph

	// The child objects that already existed prior to this processing.
	existingChildObjects serviceGraphChildObjectMaps

	// The child objects that were created during this processing.
	newChildObjects serviceGraphChildObjectMaps

	log        logr.Logger
	verboseLog logr.Logger
	setOwnerFn controllerutil.SetOwnerReferenceFn
	changes    *controllerutil.ResourceChangesList
	status     *fogappsCRDs.ServiceGraphStatus
}

// ProcessServiceGraph assembles a list of changes that need to be applied due to the specified ServiceGraph.
func ProcessServiceGraph(
	graph *fogappsCRDs.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) (*controllerutil.ResourceChangesList, *fogappsCRDs.ServiceGraphStatus, error) {
	graphProcessor := newServiceGraphProcessor(graph, childObjects, log, setOwnerFn)

	if err := graphProcessor.assembleGraphChanges(); err != nil {
		return nil, nil, err
	}

	return graphProcessor.changes, graphProcessor.status, nil
}

func newServiceGraphChildObjectMaps(lists *serviceGraphChildObjects) serviceGraphChildObjectMaps {
	maps := serviceGraphChildObjectMaps{
		Deployments:  make(map[string]*apps.Deployment),
		StatefulSets: make(map[string]*apps.StatefulSet),
		Services:     make(map[string]*core.Service),
		Ingresses:    make(map[string]*networking.Ingress),
		SloMappings:  make(map[string]*slo.UnstructuredSloMapping),
	}

	if lists != nil {
		for i := range lists.Deployments {
			// In `for i, item := rage ...`, item is apparently an object that is allocated for the loop and overwritten with
			// the values of lists.Deployments[i], which means that the address of item is the same on every iteration.
			// Using only the index in the for loop, reduces copying.
			item := &lists.Deployments[i]
			maps.Deployments[item.Name] = item
		}
		for i := range lists.StatefulSets {
			item := &lists.StatefulSets[i]
			maps.StatefulSets[item.Name] = item
		}
		for i := range lists.Services {
			item := &lists.Services[i]
			maps.Services[item.Name] = item
		}
		for i := range lists.Ingresses {
			item := &lists.Ingresses[i]
			maps.Ingresses[item.Name] = item
		}
		for i := range lists.SloMappings {
			item := &lists.SloMappings[i]
			maps.SloMappings[item.GetName()] = item
		}
	}

	return maps
}

func newServiceGraphStatus(graph *fogappsCRDs.ServiceGraph) *fogappsCRDs.ServiceGraphStatus {
	status := fogappsCRDs.ServiceGraphStatus{
		ObservedGeneration: graph.Generation,
		NodeStates:         make(map[string]*fogappsCRDs.ServiceGraphNodeStatus),
		Conditions:         make([]fogappsCRDs.ServiceGraphCondition, 0),
		SloMappings:        make([]autoscaling.CrossVersionObjectReference, 0),
	}

	// Copy the conditions from the previous Status object one-by-one to ensure that
	// we don't modify the existing Status.
	for i := range graph.Status.Conditions {
		status.Conditions = append(status.Conditions, graph.Status.Conditions[i])
	}

	return &status
}

func newServiceGraphProcessor(
	graph *fogappsCRDs.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) *serviceGraphProcessor {
	return &serviceGraphProcessor{
		svcGraph:             graph,
		existingChildObjects: newServiceGraphChildObjectMaps(childObjects),
		newChildObjects:      newServiceGraphChildObjectMaps(nil),
		log:                  log,
		verboseLog:           log.V(1),
		setOwnerFn:           setOwnerFn,
		changes:              controllerutil.NewResourceChangesList(),
		status:               newServiceGraphStatus(graph),
	}
}

// assembleGraphChanges assembles the list of changes that need to be made to align the cluster state with
// the state of the ServiceGraph.
//
// We use the following approach:
// 1. Iterate through the ServiceGraph and create or update the child objects (Deployments, StatefulSets, SLOs, etc.) and store them in newChildObjects.
// 2. Iterate through the lists of existing child objects and check if a corresponding new child object exists. We have the following options
// 		- If a corresponding new child object exists, check if there are any changes between the two specs.
//		  If there are changes, create a ResourceUpdate. In any case, remove the new child object from the newChildObjects map.
// 		- Otherwise, create a ResourceDeletion.
// 3. For all new child objects that are still in the newChildObjects map, create a ResourceAddition.
func (me *serviceGraphProcessor) assembleGraphChanges() error {
	if err := me.createChildObjectsForServiceGraph(); err != nil {
		return err
	}
	me.updateStatusConditions()

	if err := me.assembleUpdatesForDeployments(); err != nil {
		return err
	}
	if err := me.assembleUpdatesForStatefulSets(); err != nil {
		return err
	}
	if err := me.assembleUpdatesForServices(); err != nil {
		return err
	}
	if err := me.assembleUpdatesForIngresses(); err != nil {
		return err
	}
	if err := me.assembleUpdatesForSloMappings(); err != nil {
		return err
	}
	if err := me.assembleAdditions(); err != nil {
		return err
	}

	return nil
}

func (me *serviceGraphProcessor) createChildObjectsForServiceGraph() error {
	for i := range me.svcGraph.Spec.Nodes {
		node := &me.svcGraph.Spec.Nodes[i]
		switch node.NodeType {
		case fogappsCRDs.ServiceNode:
			if err := me.createChildObjectsForServiceNode(node); err != nil {
				return err
			}
		case fogappsCRDs.UserNode:
			// Nothing to be done here.
		default:
			return fmt.Errorf("unknown ServiceGraphNode.NodeType: %s", node.NodeType)
		}
	}
	return nil
}

func (me *serviceGraphProcessor) createChildObjectsForServiceNode(node *fogappsCRDs.ServiceGraphNode) error {
	var err error
	var targetRef *autoscaling.CrossVersionObjectReference

	switch node.Replicas.SetType {
	case fogappsCRDs.SimpleReplicaSet:
		targetRef, err = me.createOrUpdateDeployment(node)
	case fogappsCRDs.StatefulReplicaSet:
		targetRef, err = me.createOrUpdateStatefulSet(node)
	}

	if err != nil {
		return err
	}

	// If ExposedPorts are set, create or update the Service and Ingress.
	// If no ExposedPorts are set, not creating any Service or Ingress will cause any existing ones to be deleted later.
	if node.ExposedPorts != nil {
		if err = me.createOrUpdateServiceAndIngress(node); err != nil {
			return err
		}
	}

	// Create SloMappings from the configured SLOs, if any.
	for i := range node.SLOs {
		if err = me.createOrUpdateSloMapping(&node.SLOs[i], targetRef, node); err != nil {
			return err
		}
	}

	return nil
}

func (me *serviceGraphProcessor) createOrUpdateDeployment(node *fogappsCRDs.ServiceGraphNode) (*autoscaling.CrossVersionObjectReference, error) {
	var deployment *apps.Deployment
	var err error

	if existingDeployment, isUpdate := me.existingChildObjects.Deployments[node.Name]; isUpdate {
		deployment, err = svcGraphUtil.UpdateDeployment(existingDeployment.DeepCopy(), node, me.svcGraph)
	} else {
		if deployment, err = svcGraphUtil.CreateDeployment(node, me.svcGraph); err != nil {
			return nil, err
		}
		err = me.setOwner(deployment)
	}

	if err != nil {
		return nil, err
	}

	kubeutil.SetSpecHash(deployment, deployment.Spec)
	me.newChildObjects.Deployments[deployment.Name] = deployment
	me.updateNodeStatusWithDeployment(node, deployment)

	apiVersion, kind := deployment.GroupVersionKind().ToAPIVersionAndKind()
	targetRef := autoscaling.CrossVersionObjectReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       deployment.Name,
	}
	return &targetRef, nil
}

func (me *serviceGraphProcessor) createOrUpdateStatefulSet(node *fogappsCRDs.ServiceGraphNode) (*autoscaling.CrossVersionObjectReference, error) {
	var statefulSet *apps.StatefulSet
	var err error

	if existingStatefulSet, isUpdate := me.existingChildObjects.StatefulSets[node.Name]; isUpdate {
		statefulSet, err = svcGraphUtil.UpdateStatefulSet(existingStatefulSet.DeepCopy(), node, me.svcGraph)
	} else {
		if statefulSet, err = svcGraphUtil.CreateStatefulSet(node, me.svcGraph); err != nil {
			return nil, err
		}
		err = me.setOwner(statefulSet)
	}

	if err != nil {
		return nil, err
	}

	kubeutil.SetSpecHash(statefulSet, statefulSet.Spec)
	me.newChildObjects.StatefulSets[statefulSet.Name] = statefulSet
	me.updateNodeStatusWithStatefulSet(node, statefulSet)

	apiVersion, kind := statefulSet.GroupVersionKind().ToAPIVersionAndKind()
	targetRef := autoscaling.CrossVersionObjectReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       statefulSet.Name,
	}
	return &targetRef, nil
}

func (me *serviceGraphProcessor) createOrUpdateServiceAndIngress(node *fogappsCRDs.ServiceGraphNode) error {
	existingServiceAndIngress := &svcGraphUtil.ServiceAndIngressPair{}
	var serviceAndIngress *svcGraphUtil.ServiceAndIngressPair
	var err error

	// Check if we have an existing Service and Ingress.
	if existingService, ok := me.existingChildObjects.Services[node.Name]; ok {
		existingServiceAndIngress.Service = existingService.DeepCopy()
	}
	if existingIngress, ok := me.existingChildObjects.Ingresses[node.Name]; ok {
		existingServiceAndIngress.Ingress = existingIngress.DeepCopy()
	}

	if existingServiceAndIngress.Service != nil || existingServiceAndIngress.Ingress != nil {
		serviceAndIngress, err = svcGraphUtil.UpdateServiceAndIngress(existingServiceAndIngress, node, me.svcGraph)
	} else {
		if serviceAndIngress, err = svcGraphUtil.CreateServiceAndIngress(node, me.svcGraph); err == nil {
			if serviceAndIngress.Service != nil {
				if err = me.setOwner(serviceAndIngress.Service); err != nil {
					return err
				}
			}
			if serviceAndIngress.Ingress != nil {
				if err := me.setOwner(serviceAndIngress.Ingress); err != nil {
					return err
				}
			}
		}
	}

	if err != nil {
		return err
	}

	if serviceAndIngress.Service != nil {
		kubeutil.SetSpecHash(serviceAndIngress.Service, serviceAndIngress.Service.Spec)
		me.newChildObjects.Services[serviceAndIngress.Service.Name] = serviceAndIngress.Service
	}
	if serviceAndIngress.Ingress != nil {
		kubeutil.SetSpecHash(serviceAndIngress.Ingress, serviceAndIngress.Ingress.Spec)
		me.newChildObjects.Ingresses[serviceAndIngress.Ingress.Name] = serviceAndIngress.Ingress
	}

	return nil
}

func (me *serviceGraphProcessor) createOrUpdateSloMapping(
	sloObj *fogappsCRDs.ServiceLevelObjective,
	target *autoscaling.CrossVersionObjectReference,
	node *fogappsCRDs.ServiceGraphNode,
) error {
	newSloMapping := slo.CreateSloMappingFromServiceGraphNode(sloObj, target, node, me.svcGraph)
	me.setOwner(newSloMapping)
	kubeutil.SetSpecHash(newSloMapping, newSloMapping.Spec)
	newSloMappingUnstructured := newSloMapping.ToUnstructured()

	if existingSloMapping, ok := me.existingChildObjects.SloMappings[newSloMapping.Name]; ok {
		newSloMappingUnstructured.MergePreviousMetadata(existingSloMapping.GetMetadata())
	}

	me.newChildObjects.SloMappings[newSloMapping.Name] = newSloMappingUnstructured
	me.status.SloMappings = append(me.status.SloMappings, newSloMappingUnstructured.GetObjectReference())
	return nil
}

func (me *serviceGraphProcessor) setOwner(childObj client.Object) error {
	if err := me.setOwnerFn(childObj); err != nil {
		return fmt.Errorf("could not set owner reference. Cause: %w", err)
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForDeployments() error {
	for _, existingDeployment := range me.existingChildObjects.Deployments {
		if updatedDeployment, ok := me.newChildObjects.Deployments[existingDeployment.Name]; ok {

			if !kubeutil.CheckSpecHashesAreEqual(existingDeployment, updatedDeployment) {
				// Deployment was changed, we need to update it
				me.verboseLog.Info("Queuing update for Deployment", "deployment", updatedDeployment.Name)
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedDeployment))
			}

			delete(me.newChildObjects.Deployments, updatedDeployment.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Deployment
			me.verboseLog.Info("Queuing deletion of Deployment", "deployment", existingDeployment.Name)
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingDeployment))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForStatefulSets() error {
	for _, existingStatefulSet := range me.existingChildObjects.StatefulSets {
		if updatedStatefulSet, ok := me.newChildObjects.StatefulSets[existingStatefulSet.Name]; ok {

			if !kubeutil.CheckSpecHashesAreEqual(existingStatefulSet, updatedStatefulSet) {
				// StatefulSet was changed, we need to update it
				me.verboseLog.Info("Queuing update for StatefulSet", "statefulSet", updatedStatefulSet.Name)
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedStatefulSet))
			}

			delete(me.newChildObjects.StatefulSets, updatedStatefulSet.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the StatefulSet
			me.verboseLog.Info("Queuing deletion of StatefulSet", "statefulSet", existingStatefulSet.Name)
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingStatefulSet))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForServices() error {
	for _, existingService := range me.existingChildObjects.Services {
		if updatedService, ok := me.newChildObjects.Services[existingService.Name]; ok {

			if !kubeutil.CheckSpecHashesAreEqual(existingService, updatedService) {
				// Service was changed, we need to update it
				me.verboseLog.Info("Queuing update for Service", "service", updatedService.Name)
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedService))
			}

			delete(me.newChildObjects.Services, updatedService.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Service
			me.verboseLog.Info("Queuing deletion of Service", "service", existingService.Name)
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingService))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForIngresses() error {
	for _, existingIngress := range me.existingChildObjects.Ingresses {
		if updatedIngress, ok := me.newChildObjects.Ingresses[existingIngress.Name]; ok {

			if !kubeutil.CheckSpecHashesAreEqual(existingIngress, updatedIngress) {
				// Ingress was changed, we need to update it
				me.verboseLog.Info("Queuing update for Ingress", "ingress", updatedIngress.Name)
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedIngress))
			}

			delete(me.newChildObjects.Ingresses, updatedIngress.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Ingress
			me.verboseLog.Info("Queuing deletion of Ingress", "ingress", existingIngress.Name)
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingIngress))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForSloMappings() error {
	for _, existingSloMapping := range me.existingChildObjects.SloMappings {
		if updatedSloMapping, ok := me.newChildObjects.SloMappings[existingSloMapping.GetName()]; ok {

			if !kubeutil.CheckSpecHashesAreEqual(existingSloMapping, updatedSloMapping) {
				// SloMapping was changed, we need to update it
				me.verboseLog.Info("Queuing update for SloMapping", "sloMapping", updatedSloMapping.GetName())
				me.changes.AddChanges(controllerutil.NewResourceUpdate(&updatedSloMapping.Unstructured))
			}

			delete(me.newChildObjects.SloMappings, updatedSloMapping.GetName())
		} else {
			// The corresponding SLO or its ServiceGraphNode or ServiceGraphLink was deleted, so we delete the SloMapping
			me.verboseLog.Info("Queuing deletion of SloMapping", "sloMapping", existingSloMapping.GetName())
			me.changes.AddChanges(controllerutil.NewResourceDeletion(&existingSloMapping.Unstructured))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleAdditions() error {
	for _, value := range me.newChildObjects.Deployments {
		me.verboseLog.Info("Queuing addition of Deployment", "deployment", value.Name)
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.StatefulSets {
		me.verboseLog.Info("Queuing addition of StatefulSet", "statefulSet", value.Name)
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.Services {
		me.verboseLog.Info("Queuing addition of Service", "service", value.Name)
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.Ingresses {
		me.verboseLog.Info("Queuing addition of Ingress", "ingress", value.Name)
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.SloMappings {
		me.verboseLog.Info("Queuing addition of SloMapping", "sloMapping", value.GetName())
		me.changes.AddChanges(controllerutil.NewResourceAddition(&value.Unstructured))
	}
	return nil
}

func (me *serviceGraphProcessor) getOrCreateServiceGraphNodeStatus(node *fogappsCRDs.ServiceGraphNode) *fogappsCRDs.ServiceGraphNodeStatus {
	if nodeStatus, ok := me.status.NodeStates[node.Name]; ok {
		return nodeStatus
	}
	nodeStatus := fogappsCRDs.ServiceGraphNodeStatus{}
	me.status.NodeStates[node.Name] = &nodeStatus
	return &nodeStatus
}

func (me *serviceGraphProcessor) updateNodeStatusWithDeployment(node *fogappsCRDs.ServiceGraphNode, deployment *apps.Deployment) {
	nodeStatus := me.getOrCreateServiceGraphNodeStatus(node)
	gvk := deployment.GroupVersionKind()
	nodeStatus.DeploymentResourceType = &meta.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}

	nodeStatus.InitialReplicas = svcGraphUtil.GetInitialReplicas(node)
	if replicas := deployment.Spec.Replicas; replicas != nil {
		nodeStatus.ConfiguredReplicas = *replicas
	}
	nodeStatus.ReadyReplicas = deployment.Status.ReadyReplicas
}

func (me *serviceGraphProcessor) updateNodeStatusWithStatefulSet(node *fogappsCRDs.ServiceGraphNode, statefulSet *apps.StatefulSet) {
	nodeStatus := me.getOrCreateServiceGraphNodeStatus(node)
	gvk := statefulSet.GroupVersionKind()
	nodeStatus.DeploymentResourceType = &meta.GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}

	nodeStatus.InitialReplicas = svcGraphUtil.GetInitialReplicas(node)
	if replicas := statefulSet.Spec.Replicas; replicas != nil {
		nodeStatus.ConfiguredReplicas = *replicas
	}
	nodeStatus.ReadyReplicas = statefulSet.Status.ReadyReplicas
}

func (me *serviceGraphProcessor) updateStatusConditions() {
	isReady := true
	for i := range me.svcGraph.Spec.Nodes {
		node := &me.svcGraph.Spec.Nodes[i]
		nodeStatus := me.status.NodeStates[node.Name]
		isReady = nodeStatus.ReadyReplicas >= node.Replicas.Min
		if !isReady {
			break
		}
	}

	newCondition := fogappsCRDs.ServiceGraphCondition{
		Status:             core.ConditionTrue,
		LastTransitionTime: meta.Now(),
	}

	if isReady {
		message := "The minimum number of replicas is in a ready state for each ServiceGraphNode"
		newCondition.Type = fogappsCRDs.ServiceGraphReady
		newCondition.Reason = "MinReplicasReady"
		newCondition.Message = &message
	} else {
		message := "The minimum number of replicas is not yet in a ready state for each ServiceGraphNode"
		newCondition.Type = fogappsCRDs.ServiceGraphProgressing
		newCondition.Reason = "MinReplicasNotReady"
		newCondition.Message = &message
	}

	if len(me.status.Conditions) > 0 {
		lastCondition := me.status.Conditions[0]
		if lastCondition.Type == newCondition.Type && lastCondition.Status == newCondition.Status {
			return
		}
	}
	me.status.Conditions = []fogappsCRDs.ServiceGraphCondition{newCondition}
}
