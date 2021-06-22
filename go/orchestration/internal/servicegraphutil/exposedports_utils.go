package servicegraphutil

import (
	"fmt"

	core "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

// ServiceAndIngressPair contains the Service for a ServiceGraphNode and Ingress, if present.
type ServiceAndIngressPair struct {
	Service *core.Service
	Ingress *networking.Ingress
}

// CreateServiceAndIngress creates a new ServiceAndIngressPair, for the specified ServiceGraphNode.
func CreateServiceAndIngress(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*ServiceAndIngressPair, error) {
	ret := ServiceAndIngressPair{}
	if node.ExposedPorts == nil {
		return nil, fmt.Errorf("cannot create ServiceAndIngressPair, because node.ExposedPorts is nil")
	}

	ret.Service = createService(node, graph)

	if node.ExposedPorts.Type == fogappsCRDs.PortExposureIngress {
		ret.Ingress = createIngress(node, graph, ret.Service)
	}

	return &ret, nil
}

// UpdateServiceAndIngress updates an existing ServiceAndIngressPair for the specified ServiceGraphNode.
func UpdateServiceAndIngress(serviceAndIngress *ServiceAndIngressPair, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) (*ServiceAndIngressPair, error) {
	if node.ExposedPorts == nil {
		serviceAndIngress.Service = nil
		serviceAndIngress.Ingress = nil
		return serviceAndIngress, nil
	}

	serviceAndIngress.Service = updateService(serviceAndIngress.Service, node, graph)

	if node.ExposedPorts.Type == fogappsCRDs.PortExposureIngress {
		if serviceAndIngress.Ingress != nil {
			serviceAndIngress.Ingress = updateIngress(serviceAndIngress.Ingress, node, graph, serviceAndIngress.Service)
		} else {
			serviceAndIngress.Ingress = createIngress(node, graph, serviceAndIngress.Service)
		}
	} else {
		// Ingress is not desired, so we need to delete any existing ingress
		serviceAndIngress.Ingress = nil
	}

	return serviceAndIngress, nil
}

func createService(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *core.Service {
	service := core.Service{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       core.ServiceSpec{},
	}

	return updateService(&service, node, graph)
}

func updateService(service *core.Service, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph) *core.Service {
	var serviceType core.ServiceType
	if node.ExposedPorts.Type == fogappsCRDs.PortExposureNodeExternal {
		serviceType = core.ServiceTypeNodePort
	} else {
		serviceType = core.ServiceTypeClusterIP
	}

	updateNodeObjectMeta(&service.ObjectMeta, node, graph)
	service.Spec.Selector = getPodLabels(node, graph)
	service.Spec.Type = serviceType
	service.Spec.Ports = node.ExposedPorts.Ports

	return service
}

func createIngress(node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph, service *core.Service) *networking.Ingress {
	ingress := networking.Ingress{
		ObjectMeta: *createNodeObjectMeta(node, graph),
		Spec:       networking.IngressSpec{},
	}

	return updateIngress(&ingress, node, graph, service)
}

func updateIngress(ingress *networking.Ingress, node *fogappsCRDs.ServiceGraphNode, graph *fogappsCRDs.ServiceGraph, service *core.Service) *networking.Ingress {
	defaultPort := service.Spec.Ports[0]

	updateNodeObjectMeta(&ingress.ObjectMeta, node, graph)
	ingress.Spec.DefaultBackend = &networking.IngressBackend{
		Service: &networking.IngressServiceBackend{
			Name: service.Name,
			Port: networking.ServiceBackendPort{
				Number: defaultPort.Port,
			},
		},
	}
	ingress.Spec.Rules = nil

	return ingress
}
