package slo

import (
	"fmt"

	autoscaling "k8s.io/api/autoscaling/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	_sloMapping *SloMapping
	_           client.Object = _sloMapping
)

// SloMapping is a generic object for handling existing CRDs.
// So, we tell kubebuilder to not generate a CRD for SloMapping.
//+kubebuilder:skip

// SloMapping is a generic Kubernetes object that contains an SloMapping
type SloMapping struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty"`

	Spec   SloMappingSpec               `json:"spec,omitempty"`
	Status *fogappsCRDs.ArbitraryObject `json:"status,omitempty"`
}

// SloMappingSpec is the Spec of an SloMapping object.
type SloMappingSpec struct {

	// Specifies the target on which to execute the elasticity strategy.
	TargetRef autoscaling.CrossVersionObjectReference `json:"targetRef"`

	// The user modifiable parts of the SLO configuration from the ServiceGraph.
	fogappsCRDs.SloUserConfig `json:",inline"`
}

func (me *SloMapping) DeepCopyObject() runtime.Object {
	return &SloMapping{
		TypeMeta:   me.TypeMeta,
		ObjectMeta: me.ObjectMeta,
		Spec: SloMappingSpec{
			TargetRef:     me.Spec.TargetRef,
			SloUserConfig: *me.Spec.SloUserConfig.DeepCopy(),
		},
		Status: me.Status,
	}
}

// CreateSloMappingFromServiceGraphNode creates a new SloMapping from a service graph node.
func CreateSloMappingFromServiceGraphNode(
	slo *fogappsCRDs.ServiceLevelObjective,
	target *autoscaling.CrossVersionObjectReference,
	node *fogappsCRDs.ServiceGraphNode,
	graph *fogappsCRDs.ServiceGraph,
) *SloMapping {
	sloMapping := SloMapping{
		TypeMeta: meta.TypeMeta{
			APIVersion: slo.SloType.APIVersion,
			Kind:       slo.SloType.Kind,
		},
		ObjectMeta: meta.ObjectMeta{
			Namespace: graph.GetNamespace(),
			Name:      getSloMappingName(node.Name, slo.Name),
		},
		Spec: SloMappingSpec{
			TargetRef:     *target,
			SloUserConfig: *slo.SloUserConfig.DeepCopy(),
		},
	}
	return &sloMapping
}

func getSloMappingName(targetName string, sloName string) string {
	return fmt.Sprintf("%s-%s", targetName, sloName)
}
