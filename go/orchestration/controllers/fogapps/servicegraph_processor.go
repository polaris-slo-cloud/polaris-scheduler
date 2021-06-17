package fogapps

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	fogapps "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	svcGraphUtil "k8s.rainbow-h2020.eu/rainbow/orchestration/internal/servicegraphutil"
	"k8s.rainbow-h2020.eu/rainbow/orchestration/pkg/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceGraphChildObjectMaps struct {
	Deployments  map[string]*apps.Deployment
	StatefulSets map[string]*apps.StatefulSet
	Services     map[string]*core.Service
	Ingresses    map[string]*networking.Ingress
}

type serviceGraphProcessor struct {
	svcGraph *fogapps.ServiceGraph

	// The child objects that already existed prior to this processing.
	existingChildObjects *serviceGraphChildObjects

	// The child objects that were created during this processing.
	newChildObjects serviceGraphChildObjectMaps

	log        logr.Logger
	setOwnerFn controllerutil.SetOwnerReferenceFn
	changes    *controllerutil.ResourceChangesList
}

// ProcessServiceGraph assembles a list of changes that need to be applied due to the specified ServiceGraph.
func ProcessServiceGraph(
	graph *fogapps.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) (*controllerutil.ResourceChangesList, error) {
	graphProcessor := newServiceGraphProcessor(graph, childObjects, log, setOwnerFn)

	if err := graphProcessor.assembleGraphChanges(); err != nil {
		return nil, err
	}

	return graphProcessor.changes, nil
}

func newServiceGraphChildObjectMaps() serviceGraphChildObjectMaps {
	return serviceGraphChildObjectMaps{
		Deployments:  make(map[string]*apps.Deployment),
		StatefulSets: make(map[string]*apps.StatefulSet),
		Services:     make(map[string]*core.Service),
		Ingresses:    make(map[string]*networking.Ingress),
	}
}

func newServiceGraphProcessor(
	graph *fogapps.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) *serviceGraphProcessor {
	return &serviceGraphProcessor{
		svcGraph:             graph,
		existingChildObjects: childObjects,
		newChildObjects:      newServiceGraphChildObjectMaps(),
		log:                  log,
		setOwnerFn:           setOwnerFn,
		changes:              controllerutil.NewResourceChangesList(),
	}
}

// assembleGraphChanges assembles the list of changes that need to be made to align the cluster state with
// the state of the ServiceGraph.
//
// We use the following approach:
// 1. Iterate through the ServiceGraph and create the child objects (Deployments, StatefulSets, SLOs, etc.), as if the graph were new.
// 2. Iterate through the lists of existing child objects and check if a corresponding new child object exists. We have the following options
// 		- If a corresponding new child object exists, check if there are any changes between the two specs.
//		  If there are changes, create a ResourceUpdate. In any case, remove the new child object from the newChildObjects map.
// 		- Otherwise, create a ResourceDeletion.
// 3. For all new child objects that are still in the newChildObjects map, create a ResourceAddition.
func (me *serviceGraphProcessor) assembleGraphChanges() error {
	if err := me.createChildObjectsForServiceGraph(); err != nil {
		return err
	}

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
	if err := me.assembleAdditions(); err != nil {
		return err
	}

	return nil
}

func (me *serviceGraphProcessor) createChildObjectsForServiceGraph() error {
	for _, node := range me.svcGraph.Spec.Nodes {
		switch node.NodeType {
		case fogapps.ServiceNode:
			if err := me.createChildObjectsForServiceNode(&node); err != nil {
				return err
			}
		case fogapps.UserNode:
			// ToDo
		default:
			return fmt.Errorf("unknown ServiceGraphNode.NodeType: %s", node.NodeType)
		}
	}
	return nil
}

func (me *serviceGraphProcessor) createChildObjectsForServiceNode(node *fogapps.ServiceGraphNode) error {
	var newObj client.Object
	var deployment *apps.Deployment
	var statefulSet *apps.StatefulSet
	var err error

	switch node.Replicas.SetType {
	case fogapps.SimpleReplicaSet:
		deployment, err = svcGraphUtil.CreateDeployment(node, me.svcGraph)
		newObj = deployment
	case fogapps.StatefulReplicaSet:
		statefulSet, err = svcGraphUtil.CreateStatefulSet(node, me.svcGraph)
		newObj = statefulSet
	}

	if err != nil {
		return err
	}
	if err := me.setOwner(newObj); err != nil {
		return err
	}

	if deployment != nil {
		me.newChildObjects.Deployments[deployment.Name] = deployment
	} else if statefulSet != nil {
		me.newChildObjects.StatefulSets[statefulSet.Name] = statefulSet
	}

	if node.ExposedPorts != nil {
		if serviceAndIngress, err := svcGraphUtil.CreateServiceAndIngress(node, me.svcGraph); err == nil {
			if serviceAndIngress.Service != nil {
				if err := me.setOwner(serviceAndIngress.Service); err != nil {
					return err
				}
				me.newChildObjects.Services[serviceAndIngress.Service.Name] = serviceAndIngress.Service
			}
			if serviceAndIngress.Ingress != nil {
				if err := me.setOwner(serviceAndIngress.Ingress); err != nil {
					return err
				}
				me.newChildObjects.Ingresses[serviceAndIngress.Ingress.Name] = serviceAndIngress.Ingress
			}
		} else {
			return err
		}
	}

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
		if newDeployment, ok := me.newChildObjects.Deployments[existingDeployment.Name]; ok {

			if !reflect.DeepEqual(existingDeployment.Spec, newDeployment.Spec) {
				// Deployment was changed, we need to update it
				newDeployment.ObjectMeta = existingDeployment.ObjectMeta
				me.changes.AddChanges(controllerutil.NewResourceUpdate(newDeployment))
			}

			delete(me.newChildObjects.Deployments, newDeployment.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Deployment
			me.changes.AddChanges(controllerutil.NewResourceDeletion(&existingDeployment))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForStatefulSets() error {
	for _, existingStatefulSet := range me.existingChildObjects.StatefulSets {
		if newStatefulSet, ok := me.newChildObjects.StatefulSets[existingStatefulSet.Name]; ok {

			if !reflect.DeepEqual(existingStatefulSet.Spec, newStatefulSet.Spec) {
				// StatefulSet was changed, we need to update it
				newStatefulSet.ObjectMeta = existingStatefulSet.ObjectMeta
				me.changes.AddChanges(controllerutil.NewResourceUpdate(newStatefulSet))
			}

			delete(me.newChildObjects.StatefulSets, newStatefulSet.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the StatefulSet
			me.changes.AddChanges(controllerutil.NewResourceDeletion(&existingStatefulSet))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForServices() error {
	for _, existingService := range me.existingChildObjects.Services {
		if newService, ok := me.newChildObjects.Services[existingService.Name]; ok {

			if !reflect.DeepEqual(existingService.Spec, newService.Spec) {
				// Service was changed, we need to update it
				newService.ObjectMeta = existingService.ObjectMeta
				me.changes.AddChanges(controllerutil.NewResourceUpdate(newService))
			}

			delete(me.newChildObjects.Services, newService.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Service
			me.changes.AddChanges(controllerutil.NewResourceDeletion(&existingService))
		}
	}
	return nil
}

func (me *serviceGraphProcessor) assembleUpdatesForIngresses() error {
	for _, existingIngress := range me.existingChildObjects.Ingresses {
		if newIngress, ok := me.newChildObjects.Ingresses[existingIngress.Name]; ok {

			if !reflect.DeepEqual(existingIngress.Spec, newIngress.Spec) {
				// Ingress was changed, we need to update it
				newIngress.ObjectMeta = existingIngress.ObjectMeta
				me.changes.AddChanges(controllerutil.NewResourceUpdate(newIngress))
			}

			delete(me.newChildObjects.Ingresses, newIngress.Name)
		} else {
			// The corresponding ServiceGraphNode was deleted, so we delete the Ingress
			me.changes.AddChanges(controllerutil.NewResourceDeletion(&existingIngress))
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
	return nil
}
