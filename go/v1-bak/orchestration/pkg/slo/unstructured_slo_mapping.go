package slo

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// UnstructuredSloMapping represents an SloMapping stored as a set of nested maps.
type UnstructuredSloMapping struct {
	unstructured.Unstructured
}

// NewUnstructuredSloMapping creates a new UnstructuredSloMapping from the specified map.
func NewUnstructuredSloMapping(obj map[string]interface{}) *UnstructuredSloMapping {
	return &UnstructuredSloMapping{
		Unstructured: unstructured.Unstructured{
			Object: obj,
		},
	}
}

// GetMetadata gets the metadata object map on this UnstructuredSloMapping.
func (me *UnstructuredSloMapping) GetMetadata() map[string]interface{} {
	return me.getObject("metadata")
}

// SetMetadata sets the metadata object on this UnstructuredSloMapping.
func (me *UnstructuredSloMapping) SetMetadata(metadata map[string]interface{}) {
	me.Object["metadata"] = metadata
}

// Merges relevant metadata (needed for updating an object) from a previous version of this SLO into this version's metadata.
func (me *UnstructuredSloMapping) MergePreviousMetadata(prevMetadata map[string]interface{}) {
	metadata := me.Object["metadata"].(map[string]interface{})

	mergeField := func(key string) {
		if value, found := prevMetadata[key]; found {
			metadata[key] = value
		}
	}

	mergeField("resourceVersion")
	mergeField("uid")
}

// GetSpec gets the spec object map on this UnstructuredSloMapping.
func (me *UnstructuredSloMapping) GetSpec() map[string]interface{} {
	return me.getObject("spec")
}

// DeleteStatus deletes the status object of this UnstructuredSloMapping.
func (me *UnstructuredSloMapping) DeleteStatus() {
	delete(me.Object, "status")
}

// GetObjectReference creates a CrossVersionObjectReference for this SloMapping.
func (me *UnstructuredSloMapping) GetObjectReference() autoscaling.CrossVersionObjectReference {
	return autoscaling.CrossVersionObjectReference{
		APIVersion: me.GetAPIVersion(),
		Kind:       me.GetKind(),
		Name:       me.GetName(),
	}
}

func (me *UnstructuredSloMapping) DeepCopyObject() runtime.Object {
	return &UnstructuredSloMapping{
		Unstructured: *me.Unstructured.DeepCopy(),
	}
}

func (me *UnstructuredSloMapping) getObject(key string) map[string]interface{} {
	if metadata, ok := me.Object[key]; ok {
		return metadata.(map[string]interface{})
	}
	return nil
}
