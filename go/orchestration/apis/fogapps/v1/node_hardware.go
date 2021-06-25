package v1

// CpuArchitecture defines the possible CPU architectures.
//
// +kubebuilder:validation:Enum=386;amd64;arm;arm64
type CpuArchitecture string

// The list of CPU architecures is a selection taken from https://gist.github.com/asukakenji/f15ba7e588ac42795f421b48b8aede63
var (
	// Intel 32-bit CPU architecture
	CpuArch386 CpuArchitecture = "386"

	// Intel 64-bit CPU architecture
	CpuArchAmd64 CpuArchitecture = "amd64"

	// ARM 32-bit CPU architecture
	CpuArchArm CpuArchitecture = "arm"

	// ARM 64-bit CPU architecture
	CpuArchArm64 CpuArchitecture = "arm64"
)

// NodeHardware is used to specify hardware requirements for cluster nodes hosting an instance of a ServiceGraphNode.
//
// These requirements define only what hardware capabilities the hosting node needs to have, but do not
// request exclusive access for them (if necessary, this can be done in the resources configuration of a Container).
type NodeHardware struct {

	// A string that allows specifying the type of host node, e.g., a cloud server,
	// a stationary compute node in the fog, or an airborne drone in the fog.
	//
	// A node type string is hierarchical and composed similar to a URI:
	// <fog|cloud>/<category>/<optionalvendor>/<type>/<optional-model>
	//
	// Each hierarchy level designates a set of compliant nodes, which can be reduced by specifying an additional level.
	// For example, the node type string “fog/stationary” ensures that the service is only deployed on
	// stationary fog nodes, i.e., a road side unit integrated into a traffic light would be eligible,
	// but a smart car would not.
	// The more specific string “fog/stationary/raspberrypi/4-b” would only allow fog nodes that are a stationary Raspberry Pi Model 4 B.
	//
	// +optional
	NodeType *string `json:"nodeType,omitempty"`

	// ToDo Add NodeType validation regex with kubebuilder:validation:Pattern

	// Used to define CPU requirements for the host node, e.g., CPU architecture.
	//
	// +optional
	CpuInfo *CpuInfo `json:"cpuInfo,omitempty"`

	// Used to define GPU requirements for the host node.
	//
	// +optional
	GpuInfo *GpuInfo `json:"gpuInfo,omitempty"`
}

// CpuInfo describes requirements for the CPU of a host node, e.g., architecture and number of cores.
//
// All fields are optional.
type CpuInfo struct {

	// Array of CPU architectures supported by the container images of this service.
	// If this is not specified, all CPU architectures are assumed to be supported.
	//
	// Possible CpuArchitecture values:
	// - Intel 32-bit: "386"
	// - Intel 64-bit: "amd64"
	// - ARM 32-bit: "arm"
	// - ARM 64-bit: "arm64"
	//
	// +optional
	Architectures []CpuArchitecture `json:"architectures,omitempty"`

	// The minimum number of CPU cores that the node must have.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinCores *int32 `json:"minCores,omitempty"`

	// The minimum base clock frequency of the CPU in MHz.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinBashClockMHz *int32 `json:"minBashClockMHz,omitempty"`
}

// GpuInfo describes requirements for the GPU of a host node.
//
// All fields are optional.
type GpuInfo struct {

	// ToDo
}
