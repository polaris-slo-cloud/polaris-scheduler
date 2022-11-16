package v1

// ArbitraryObject is used to model an object with unknown contents in a CRD.
//
// It can, e.g., be used to store SLO configuration data, which depends on the SLO type
// and cannot be known by the ServiceGraph CRD.
//
// Note that the Go deepCopy() method will only be able to create shallow copies of this field,
// because it does not know its internal structure.
//
// +kubebuilder:pruning:PreserveUnknownFields
type ArbitraryObject struct{}
