package v1

import (
	cluster "k8s.rainbow-h2020.eu/rainbow/apis/cluster/v1"
)

// LinkProtocol is used to describe the type of protocol that will be used for the communication over a ServiceLink.
//
// +kubebuilder:validation:Enum=http;https;tcp;udp
type LinkProtocol string

var (
	HttpProtocol  LinkProtocol = "http"
	HttpsProtocol LinkProtocol = "https"
	TcpProtocol   LinkProtocol = "tcp"
	UdpProtocol   LinkProtocol = "udp"
)

// LinkType describes requirements for the type of network link that a ServiceLink needs.
type LinkType struct {

	// The type of protocol that will be used for the communication over a ServiceLink.
	//
	// +optional
	Protocol *LinkProtocol `json:"protocol,omitempty"`

	// The required minimum quality class of this network link.
	//
	// +optional
	MinQualityClass *cluster.NetworkQualityClass `json:"minQualityClass"`
}
