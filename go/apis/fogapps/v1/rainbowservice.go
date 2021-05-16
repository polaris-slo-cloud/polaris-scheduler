package v1

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RainbowService describes the configuration of a RAINBOW platform service.
type RainbowService struct {

	// Defines the type of RAINBOW service.
	ServiceType meta.GroupVersionKind `json:"serviceType"`

	// The service-specific configuration.
	Config *ArbitraryObject `json:"config,omitempty"`
}
