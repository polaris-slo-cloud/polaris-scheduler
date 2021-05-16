package v1

// RainbowService describes the configuration of a RAINBOW platform service.
type RainbowService struct {

	// Defines the type of RAINBOW service.
	ServiceType ApiVersionKind `json:"serviceType"`

	// The service-specific configuration.
	Config *ArbitraryObject `json:"config,omitempty"`
}
