package v1

// ConfigMap is like the Kubernetes ConfigMap type, except that it is designed to
// be embedded into a ServiceGraph object.
//
// The code is a modified version of https://pkg.go.dev/k8s.io/api/core/v1#ConfigMap
type ConfigMap struct {

	// Immutable, if set to true, ensures that data stored in the ConfigMap cannot be updated.
	// If not set to true, the field can be modified at any time.
	//
	// +kubebuilder:default=false
	// +optional
	Immutable bool `json:"immutable"`

	// Data contains the configuration data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// Values with non-UTF-8 byte sequences must use the BinaryData field.
	// The keys stored in Data must not overlap with the keys in
	// the BinaryData field, this is enforced during validation process.
	//
	// +optional
	Data map[string]string `json:"data,omitempty"`

	// BinaryData contains the binary data.
	// Each key must consist of alphanumeric characters, '-', '_' or '.'.
	// BinaryData can contain byte sequences that are not in the UTF-8 range.
	// The keys stored in BinaryData must not overlap with the ones in
	// the Data field, this is enforced during validation process.
	//
	// +optional
	BinaryData map[string][]byte `json:"binaryData,omitempty"`
}
