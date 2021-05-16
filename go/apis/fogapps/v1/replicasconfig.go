package v1

// ReplicaSetType defines the available types of replica sets.
//
// +kubebuilder:validation:Enum=Simple;Stateful
type ReplicaSetType string

var (
	SimpleReplicaSet   ReplicaSetType = "Simple"
	StatefulReplicaSet ReplicaSetType = "Stateful"
)

// ReplicasConfig specifies the minimum, maximum, and initial replica count,
// as well as the type of replica set.
type ReplicasConfig struct {

	// The minium number of replicas (default = 1).
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	Min int32 `json:"min,omitempty"`

	// The maximum number of replicas.
	//
	// +kubebuilder:validation:Minimum=1
	Max int32 `json:"max"`

	// The initial number of replicas that should be created upon deployment.
	// Defaults to the same number as Min.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	InitialCount *int32 `json:"initialCount,omitempty"`

	// Specifies the type of replica set that should be used.
	//
	// The possibilities are:
	//
	// - "Simple" (default) Creates and destroys instances of the service, treating them as stateless.
	//
	// - "Stateful" Ensures that the set of replicas is ordered (i.e., replica 2 is always created before replica 3)
	// and that each specific replica is always connected to the same volumes it was originally connected to.
	//
	// +kubebuilder:default=Simple
	SetType ReplicaSetType `json:"setType,omitempty"`
}
