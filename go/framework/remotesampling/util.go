package remotesampling

import (
	"math"

	core "k8s.io/api/core/v1"
	"polaris-slo-cloud.github.io/polaris-scheduler/v2/framework/collections"
)

// Calculates the number of nodes that a sample for the specified request needs to contain.
func CalcRequiredNodesCount(
	request *RemoteNodesSamplerRequest,
	storeReader collections.ConcurrentObjectStoreReader[*core.Node],
) int {
	percentageOfNodesToSample := float64(request.NodesToSampleBp) / 10000.0
	reqNodes := percentageOfNodesToSample * float64(storeReader.Len())
	return int(math.Max(reqNodes, 1))
}
