package servicegraphutil

import (
	"fmt"

	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	fogapps "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

// ServiceAndIngressPair contains the Service for a ServiceGraphNode and Ingress, if present.
type ServiceAndIngressPair struct {
	Service *core.Service
	Ingress *networking.Ingress
}

// CreateServiceAndIngress creates a Service and Ingress, if necessary, for the specified ServiceGraphNode.
func CreateServiceAndIngress(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) (ServiceAndIngressPair, error) {
	ret := ServiceAndIngressPair{}
	if node.ExposedPorts == nil {
		return ret, fmt.Errorf("cannot create ServiceAndIngressPair, because node.ExposedPorts is nil")
	}

	ret.Service = createService(node, graph)

	if node.ExposedPorts.Type == fogapps.PortExposureIngress {
		ret.Ingress = createIngress(node, graph, ret.Service)
	}

	return ret, nil
}

func createService(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph) *core.Service {
	var serviceType core.ServiceType
	if node.ExposedPorts.Type == fogapps.PortExposureNodeExternal {
		serviceType = core.ServiceTypeNodePort
	} else {
		serviceType = core.ServiceTypeClusterIP
	}

	return &core.Service{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec: core.ServiceSpec{
			Selector: getPodLabels(node, graph),
			Type:     serviceType,
			Ports:    node.ExposedPorts.Ports,
		},
	}
}

func createIngress(node *fogapps.ServiceGraphNode, graph *fogapps.ServiceGraph, service *core.Service) *networking.Ingress {
	defaultPort := service.Spec.Ports[0]

	return &networking.Ingress{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec: networking.IngressSpec{
			DefaultBackend: &networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: service.Name,
					Port: networking.ServiceBackendPort{
						Number: defaultPort.Port,
					},
				},
			},
		},
	}
}
