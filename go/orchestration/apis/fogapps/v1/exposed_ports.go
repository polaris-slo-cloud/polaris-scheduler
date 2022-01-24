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
	// - "Ingress" Exposes the ports using a load balanced Ingress controller with the first port of this node being the
	// default backend. To configure additional rules, the IngressConfig field (ToDo) must be filled.
	//
	// +kubebuilder:default=ClusterInternal
	// +optional
	Type PortExposureType `json:"type"`

	// Configures the ports that should be exposed.
	Ports []core.ServicePort `json:"ports"`

	// ToDo:
	// IngressConfig (we need to extract a subset of https://pkg.go.dev/k8s.io/api@v0.21.0/networking/v1#IngressSpec because the ServiceName
	// should be set by the ServiceGraph controller)
	// We will probably only need the `Rules`and possibly the `TLS` fields, because the `DefaultBackend` is the ServiceGraphNode, where
	// the `ExposedPorts` are configured. For the rules, we could use the names of other service graph nodes instead of services,
	// but the services need to be created there using `ExposedPorts` of course.
	// Add this link to the docs: https://kubernetes.io/docs/concepts/services-networking/ingress/
}
