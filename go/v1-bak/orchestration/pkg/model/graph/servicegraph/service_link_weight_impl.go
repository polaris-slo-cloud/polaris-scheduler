package servicegraph

import (
	fogappsCRDs "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/fogapps/v1"
)

var (
	_weightImpl *serviceLinkWeightImpl

	_ ServiceLinkWeight = _weightImpl
)

type serviceLinkWeightImpl struct {
	serviceLink *fogappsCRDs.ServiceLink
}

func newServiceLinkWeightImpl(serviceLink *fogappsCRDs.ServiceLink) *serviceLinkWeightImpl {
	return &serviceLinkWeightImpl{
		serviceLink: serviceLink,
	}
}

func (me *serviceLinkWeightImpl) ServiceLink() *fogappsCRDs.ServiceLink {
	return me.serviceLink
}

func (me *serviceLinkWeightImpl) SimpleWeight() float64 {
	if me.serviceLink.QosRequirements != nil && me.serviceLink.QosRequirements.Latency != nil {
		return float64(me.serviceLink.QosRequirements.Latency.MaxPacketDelayMsec)
	}
	return 0
}
