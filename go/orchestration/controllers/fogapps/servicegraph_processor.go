package fogapps

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	autoscaling "k8s.io/api/autoscaling/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	svcGraphUtil "k8s.rainbow-h2020.eu/rainbow/orchestration/internal/servicegraphutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/controllerutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/slo"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceGraphChildObjectMaps struct {
	Deployments  map[string]*apps.Deployment
	StatefulSets map[string]*apps.StatefulSet
	Services     map[string]*core.Service
	Ingresses    map[string]*networking.Ingress
	SloMappings  map[string]*slo.SloMapping
}

type serviceGraphProcessor struct {
	svcGraph *fogappsCRDs.ServiceGraph

	// The child objects that already existed prior to this processing.
	existingChildObjects serviceGraphChildObjectMaps

	// The child objects that were created during this processing.
	newChildObjects serviceGraphChildObjectMaps

	log        logr.Logger
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
		SloMappings:  make(map[string]*slo.SloMapping),
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
			maps.SloMappings[item.Name] = item
		}
	}

	return maps
}

func newServiceGraphStatus(graph *fogappsCRDs.ServiceGraph) *fogappsCRDs.ServiceGraphStatus {
	status := fogappsCRDs.ServiceGraphStatus{
		ObservedGeneration: graph.Generation,
		NodeStates:         make(map[string]*fogappsCRDs.ServiceGraphNodeStatus),
		Conditions:         make([]fogappsCRDs.ServiceGraphCondition, 0),
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
	for _, node := range me.svcGraph.Spec.Nodes {
		switch node.NodeType {
		case fogappsCRDs.ServiceNode:
			if err := me.createChildObjectsForServiceNode(&node); err != nil {
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

	me.newChildObjects.Deployments[deployment.Name] = deployment
	me.updateNodeStatusWithDeployment(node, deployment)

	targetRef := autoscaling.CrossVersionObjectReference{
		APIVersion: deployment.APIVersion,
		Kind:       deployment.Kind,
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

	me.newChildObjects.StatefulSets[statefulSet.Name] = statefulSet
	me.updateNodeStatusWithStatefulSet(node, statefulSet)

	targetRef := autoscaling.CrossVersionObjectReference{
		APIVersion: statefulSet.APIVersion,
		Kind:       statefulSet.Kind,
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
		me.newChildObjects.Services[serviceAndIngress.Service.Name] = serviceAndIngress.Service
	}
	if serviceAndIngress.Ingress != nil {
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

	if existingSloMapping, ok := me.existingChildObjects.SloMappings[newSloMapping.Name]; ok {
		newSloMapping.ObjectMeta = existingSloMapping.ObjectMeta
	} else {
		me.setOwner(newSloMapping)
	}

	me.newChildObjects.SloMappings[newSloMapping.Name] = newSloMapping
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

			// ToDo: Containers are currently never equal, because Kubernetes sets some values, which are unset in new containers.
			// containersEqual := reflect.DeepEqual(existingDeployment.Spec.Template.Spec.Containers, updatedDeployment.Spec.Template.Spec.Containers)
			// _ = containersEqual

			if !reflect.DeepEqual(existingDeployment.Spec, updatedDeployment.Spec) {
				// Deployment was changed, we need to update it
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedDeployment))
			}

			delete(me.newChildObjects.Deployments, updatedDeployment.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Deployment
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingDeployment))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForStatefulSets() error {
	for _, existingStatefulSet := range me.existingChildObjects.StatefulSets {
		if updatedStatefulSet, ok := me.newChildObjects.StatefulSets[existingStatefulSet.Name]; ok {

			if !reflect.DeepEqual(existingStatefulSet.Spec, updatedStatefulSet.Spec) {
				// StatefulSet was changed, we need to update it
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedStatefulSet))
			}

			delete(me.newChildObjects.StatefulSets, updatedStatefulSet.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the StatefulSet
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingStatefulSet))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForServices() error {
	for _, existingService := range me.existingChildObjects.Services {
		if updatedService, ok := me.newChildObjects.Services[existingService.Name]; ok {

			if !reflect.DeepEqual(existingService.Spec, updatedService.Spec) {
				// Service was changed, we need to update it
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedService))
			}

			delete(me.newChildObjects.Services, updatedService.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Service
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingService))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForIngresses() error {
	for _, existingIngress := range me.existingChildObjects.Ingresses {
		if updatedIngress, ok := me.newChildObjects.Ingresses[existingIngress.Name]; ok {

			if !reflect.DeepEqual(existingIngress.Spec, updatedIngress.Spec) {
				// Ingress was changed, we need to update it
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedIngress))
			}

			delete(me.newChildObjects.Ingresses, updatedIngress.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Ingress
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingIngress))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForSloMappings() error {
	for _, existingSloMapping := range me.existingChildObjects.SloMappings {
		if updatedSloMapping, ok := me.newChildObjects.SloMappings[existingSloMapping.Name]; ok {

			if !reflect.DeepEqual(existingSloMapping.Spec, updatedSloMapping.Spec) {
				// SloMapping was changed, we need to update it
				me.changes.AddChanges(controllerutil.NewResourceUpdate(updatedSloMapping))
			}

			delete(me.newChildObjects.SloMappings, updatedSloMapping.Name)
		} else {
			// The corresponding SLO or its ServiceGraphNode or ServiceGraphLink was deleted, so we delete the SloMapping
			me.changes.AddChanges(controllerutil.NewResourceDeletion(existingSloMapping))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleAdditions() error {
	for _, value := range me.newChildObjects.Deployments {
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.StatefulSets {
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.Services {
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.Ingresses {
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
	}
	for _, value := range me.newChildObjects.SloMappings {
		me.changes.AddChanges(controllerutil.NewResourceAddition(value))
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
	nodeStatus.ConfiguredReplicas = *deployment.Spec.Replicas
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
	nodeStatus.ConfiguredReplicas = *statefulSet.Spec.Replicas
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
