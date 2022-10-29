package regiongraph

import (
	cluster "k8s.rainbow-h2020.eu/rainbow/orchestration/apis/cluster/v1"
)

var (
	_weightImpl *networkLinkQosWeightImpl

	_ NetworkLinkQosWeight = _weightImpl
)

type networkLinkQosWeightImpl struct {
	qos *cluster.NetworkLinkQoS
}

func newNetworkLinkQosWeightImpl(qos *cluster.NetworkLinkQoS) *networkLinkQosWeightImpl {
	return &networkLinkQosWeightImpl{
		qos: qos,
	}
}

func (me *networkLinkQosWeightImpl) NetworkLinkQoS() *cluster.NetworkLinkQoS {
	return me.qos
}

func (me *networkLinkQosWeightImpl) SimpleWeight() float64 {
	return float64(me.qos.Latency.PacketDelayMsec)
}
