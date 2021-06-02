package fogapps

import (
	"fmt"

	"github.com/go-logr/logr"
	fogapps "k8s.rainbow-h2020.eu/rainbow/apis/fogapps/v1"
	svcGraphUtil "k8s.rainbow-h2020.eu/rainbow/internal/servicegraphutil"
	"k8s.rainbow-h2020.eu/rainbow/pkg/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type serviceGraphProcessor struct {
	svcGraph     *fogapps.ServiceGraph
	childObjects *serviceGraphChildObjects
	log          logr.Logger
	setOwnerFn   controllerutil.SetOwnerReferenceFn
	changes      *controllerutil.ResourceChangesList
}

// ProcessServiceGraph assembles a list of changes that need to be applied due to the specified ServiceGraph.
func ProcessServiceGraph(
	graph *fogapps.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) (*controllerutil.ResourceChangesList, error) {
	graphProcessor := newServiceGraphProcessor(graph, childObjects, log, setOwnerFn)

	if err := graphProcessor.assembleNodeChanges(); err != nil {
		return nil, err
	}

	return graphProcessor.changes, nil
}

func newServiceGraphProcessor(
	graph *fogapps.ServiceGraph,
	childObjects *serviceGraphChildObjects,
	log logr.Logger,
	setOwnerFn controllerutil.SetOwnerReferenceFn,
) *serviceGraphProcessor {
	return &serviceGraphProcessor{
		svcGraph:     graph,
		childObjects: childObjects,
		log:          log,
		setOwnerFn:   setOwnerFn,
		changes:      controllerutil.NewResourceChangesList(),
	}
}

func (me *serviceGraphProcessor) assembleNodeChanges() error {
	for _, node := range me.svcGraph.Spec.Nodes {
		switch node.NodeType {
		case fogapps.ServiceNode:
			if err := me.handleServiceNode(&node); err != nil {
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

func (me *serviceGraphProcessor) handleServiceNode(node *fogapps.ServiceGraphNode) error {
	var newObj client.Object
	var err error

	switch node.Replicas.SetType {
	case fogapps.SimpleReplicaSet:
		newObj, err = svcGraphUtil.CreateDeployment(node, me.svcGraph)
	case fogapps.StatefulReplicaSet:
		newObj, err = svcGraphUtil.CreateStatefulSet(node, me.svcGraph)
	}

	if err != nil {
		return err
	}
	if err := me.setOwnerFn(newObj); err != nil {
		return fmt.Errorf("could not set owner reference. Cause: %w", err)
	}

	// Workaround until we implement updating
	if len(me.childObjects.Deployments) == 0 {
		me.changes.AddChanges(controllerutil.NewResourceAddition(newObj))
	}
	return nil
}
