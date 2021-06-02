package v1

// TpmType defines the available TPM types.
//
// +kubebuilder:validation:Enum=none;software;hardware
type TpmType string

var (
	NoTPM       TpmType = "none"
	SoftwareTPM TpmType = "software"
	HardwareTPM TpmType = "hardware"
)

// NodeTrustRequirements is used to configure the trust requirements for a ServiceGraphNode.
type NodeTrustRequirements struct {

	// The type of TPM that is needed
	TpmType TpmType `json:"tpmType"`

	// A string denoting the version of TPM that is required, e.g., "2.0".
	//
	// +optional
	MinTpmVersion *string `json:"minTpmVersion,omitempty"`

	// ToDo
	// - attestability
	// - others?
}
