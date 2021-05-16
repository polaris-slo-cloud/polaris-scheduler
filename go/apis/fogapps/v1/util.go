package v1

// ApiVersionKind allows indicating another API resource type from a CRD.
//
// This type should be used for referencing other API resource types in a CRD instead
// of meta.GroupVersionKind, because it combines the API group and version into a single field,
// thus, being more consistent with Kubernetes YAML files.
// We also cannot use meta.TypeMeta, because all its fields are optional.
type ApiVersionKind struct {
	// The API group and version of the type.
	APIVersion string `json:"apiVersion"`

	// The kind name of the type.
	Kind string `json:"kind"`
}
