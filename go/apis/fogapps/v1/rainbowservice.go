package v1

// RainbowService describes the configuration of a RAINBOW platform service.
type RainbowService struct {

	// Defines the type of RAINBOW service.
	Type ApiVersionKind `json:"type"`

	// The service-specific configuration.
	//
	// +optional
	Config *ArbitraryObject `json:"config,omitempty"`
}
