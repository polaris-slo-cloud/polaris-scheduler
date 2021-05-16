package v1

import (
	core "k8s.io/api/core/v1"
)

// PortExposureType defines in which scope the ports will be exposed.
//
// +kubebuilder:validation:Enum=ClusterInternal;NodeExternal;Ingress
type PortExposureType string

var (
	// Exposes the ports within the cluster only.
	PortExposureClusterInternal PortExposureType = "ClusterInternal"

	// Exposes each port as an externally open port on every node in the cluster.
	// Recommended only for debugging.
	PortExposureNodeExternal PortExposureType = "NodeExternal"

	// Exposes the ports using a load balanced Ingress controller.
	PortExposureIngress PortExposureType = "Ingress"
)

// ExposedPorts allows configuring ports that should be exposed by a ServiceGraphNode.
type ExposedPorts struct {

	// Defines in which scope the ports will be exposed.
	// The possibilities are:
	//
	// - "ClusterInternal" (default) Exposes the ports within the cluster only.
	//
	// - "NodeExternal" Exposes each port as an externally open port on every node in the cluster.
	// Recommended only for debugging.
	//
	// - "Ingress" Exposes the ports using a load balanced Ingress controller.
	// In this case, the IngressConfig field must be filled.
	//
	// +kubebuilder:default=ClusterInternal
	Type PortExposureType `json:"type"`

	// Configures the ports that should be exposed.
	Ports []core.ServicePort `json:"ports"`

	// ToDo:
	// IngressConfig (we need to extract a subset of https://pkg.go.dev/k8s.io/api@v0.21.0/networking/v1#IngressSpec because the ServiceName
	// should be set by the ServiceGraph controller)
}
