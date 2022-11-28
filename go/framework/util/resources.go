package util

import (
	core "k8s.io/api/core/v1"
)

// Represents a set of resources requested or consumed by a Pod.
type Resources struct {
	MilliCPU          int64                       `json:"milliCpu" yaml:"milliCpu"`
	MemoryBytes       int64                       `json:"memoryBytes" yaml:"memoryBytes"`
	EphemeralStorage  int64                       `json:"ephemeralStorage" yaml:"ephemeralStorage"`
	ExtendedResources map[core.ResourceName]int64 `json:"extendedResources" yaml:"extendedResources"`
}

// Creates a new, empty Resources object.
func NewResources() *Resources {
	return &Resources{}
}

// Creates a Resources object and initializes it with the specified ResourceList.
func NewResourcesFromList(rl core.ResourceList) *Resources {
	resources := NewResources()
	resources.AddResourceList(rl)
	return resources
}

// Adds the values in the specified Resources object to this Resources object.
func (r *Resources) Add(other *Resources) {
	r.MilliCPU += other.MilliCPU
	r.MemoryBytes += other.MemoryBytes
	r.EphemeralStorage += other.EphemeralStorage

	for resName, value := range other.ExtendedResources {
		r.addExtendedResource(resName, value)
	}
}

// Adds the specified ResourceList to this Resources object.
func (r *Resources) AddResourceList(rl core.ResourceList) {
	for name, quantity := range rl {
		switch name {
		case core.ResourceCPU:
			r.MilliCPU += quantity.MilliValue()
		case core.ResourceMemory:
			r.MemoryBytes += quantity.Value()
		case core.ResourceEphemeralStorage:
			r.EphemeralStorage += quantity.Value()
		default:
			r.addExtendedResource(name, quantity.Value())
		}
	}
}

// Subtracts the values in the specified Resources object from this Resources object.
func (r *Resources) Subtract(other *Resources) {
	r.MilliCPU -= other.MilliCPU
	r.MemoryBytes -= other.MemoryBytes
	r.EphemeralStorage -= other.EphemeralStorage

	for resName, value := range other.ExtendedResources {
		r.addExtendedResource(resName, -value)
	}
}

// Subtracts the specified ResourceList from this Resources object.
func (r *Resources) SubtractResourceList(rl core.ResourceList) {
	for name, quantity := range rl {
		switch name {
		case core.ResourceCPU:
			r.MilliCPU -= quantity.MilliValue()
		case core.ResourceMemory:
			r.MemoryBytes -= quantity.Value()
		case core.ResourceEphemeralStorage:
			r.EphemeralStorage -= quantity.Value()
		default:
			r.addExtendedResource(name, -quantity.Value())
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

// Returns true if the values in this object are the same as in the other resources object.
func (r *Resources) Equals(other *Resources) bool {
	if r.MilliCPU != other.MilliCPU {
		return false
	}
	if r.MemoryBytes != other.MemoryBytes {
		return false
	}
	if r.EphemeralStorage != other.EphemeralStorage {
		return false
	}
	if r.ExtendedResources != nil && other.ExtendedResources == nil {
		return false
	}
	if len(r.ExtendedResources) != len(other.ExtendedResources) {
		return false
	}

	for name, rQuantity := range r.ExtendedResources {
		if otherQuantity, ok := other.ExtendedResources[name]; !ok || rQuantity != otherQuantity {
			return false
		}
	}

	return true
}

// Creates a deep copy of this Resources object.
func (r *Resources) DeepCopy() *Resources {
	ret := NewResources()
	ret.Add(r)
	return ret
}

func (r *Resources) addExtendedResource(name core.ResourceName, rQuantityValue int64) {
	if r.ExtendedResources == nil {
		r.ExtendedResources = make(map[core.ResourceName]int64)
	}

	var newValue int64
	if existingValue, ok := r.ExtendedResources[name]; ok {
		newValue = existingValue + rQuantityValue
	} else {
		newValue = rQuantityValue
	}

	r.ExtendedResources[name] = newValue
}
