package util

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Represents a set of resources requested or consumed by a Pod.
type Resources struct {
	MilliCPU          int64
	MemoryBytes       int64
	EphemeralStorage  int64
	ExtendedResources map[core.ResourceName]int64
}

// Creates a new, empty Resources object.
func NewResources() *Resources {
	return &Resources{}
}

// Creates a Resources object and initializes it with the specified ResourceList.
func NewResourcesFromList(rl core.ResourceList) *Resources {
	resources := NewResources()
	resources.Add(rl)
	return resources
}

// Adds the specified ResourceList to this Resources object.
func (r *Resources) Add(rl core.ResourceList) {
	for name, quantity := range rl {
		switch name {
		case core.ResourceCPU:
			r.MilliCPU += quantity.MilliValue()
		case core.ResourceMemory:
			r.MemoryBytes += quantity.Value()
		case core.ResourceEphemeralStorage:
			r.EphemeralStorage += quantity.Value()
		default:
			r.addExtendedResource(name, quantity)
		}
	}
}

// Returns true if all resource numbers expressed by this object are less than or equal to
// the ones expressed in the other object.
func (r *Resources) LessThanOrEqual(other *Resources) bool {
	if r.MilliCPU > other.MilliCPU {
		return false
	}
	if r.MemoryBytes > other.MemoryBytes {
		return false
	}
	if r.EphemeralStorage > other.EphemeralStorage {
		return false
	}
	if r.ExtendedResources != nil && other.ExtendedResources == nil {
		return false
	}

	for name, rQuantity := range r.ExtendedResources {
		if otherQuantity, ok := other.ExtendedResources[name]; !ok || rQuantity > otherQuantity {
			return false
		}
	}

	return true
}

func (r *Resources) addExtendedResource(name core.ResourceName, quantity resource.Quantity) {
	if r.ExtendedResources == nil {
		r.ExtendedResources = make(map[core.ResourceName]int64)
	}

	var newValue int64
	if existingValue, ok := r.ExtendedResources[name]; ok {
		newValue = existingValue + quantity.Value()
	} else {
		newValue = quantity.Value()
	}

	r.ExtendedResources[name] = newValue
}
