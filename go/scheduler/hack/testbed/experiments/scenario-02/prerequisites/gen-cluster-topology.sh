#!/bin/bash

CLOUD_NODE_PREFIX="cloud-medium-"

function printUsage() {
    echo "gen-cluster-topology.sh generates and prints the NetworkLink objects for a single subcluster consisting of 12 nodes."
    echo "The cloud node is connected to the cloud nodes of the other subclusters."
    echo "Each subcluster is identified by an ID, which must be supplied as an argument."
    echo "To generate and apply the topology of a cluster that consists of 3 subclusters, run the following commands:"
    echo "./gen-cluster-topology.sh 0 | kubectl apply -f -"
    echo "./gen-cluster-topology.sh 1 | kubectl apply -f -"
    echo "./gen-cluster-topology.sh 2 | kubectl apply -f -"
}

function generateIntraCloudLinks() {
    local clusterId=$1
    let prevClusterId=clusterId-1
    local clusterCloudNode="${CLOUD_NODE_PREFIX}${clusterId}"
    local retYaml=""

    for i in $(seq 0 $prevClusterId); do
        local currCloudNode="${CLOUD_NODE_PREFIX}${i}"
        yaml=$(cat <<EOF
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${currCloudNode}-to-${clusterCloudNode}
spec:
  nodeA: ${currCloudNode}
  nodeB: ${clusterCloudNode}
  qos:
    qualityClass: QC1Gbps
    throughput:
      bandwidthKbps: 1000000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 1
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
EOF
        )
        retYaml=$(echo -e "${retYaml}\n${yaml}")
    done

    RET="${retYaml}"
}

function generateSubcluster() {
    # Node names template:
    # cloud-medium-${CLUSTER_ID}
    # raspi-3b-${CLUSTER_ID * 2}
    # raspi-3b-${CLUSTER_ID * 2 + 1}
    # raspi-4s-${CLUSTER_ID * 2}
    # raspi-4s-${CLUSTER_ID * 2 + 1}
    # raspi-4m-${CLUSTER_ID * 4}
    # raspi-4m-${CLUSTER_ID * 4 + 1}
    # raspi-4m-${CLUSTER_ID * 4 + 2}
    # raspi-4m-${CLUSTER_ID * 4 + 3}
    # base-station-5g-${CLUSTER_ID * 3}
    # base-station-5g-${CLUSTER_ID * 3 + 1}
    # base-station-5g-${CLUSTER_ID * 3 + 2}

    local clusterId=$1

    generateIntraCloudLinks $clusterId
    local intraCloudYaml="${RET}"

    local cloudNode="${CLOUD_NODE_PREFIX}${clusterId}"
    let i=clusterId*2
    local raspi3b0="raspi-3b-${i}"
    local raspi4s0="raspi-4s-${i}"
    let i++
    local raspi3b1="raspi-3b-${i}"
    local raspi4s1="raspi-4s-${i}"
    let i=clusterId*4
    local raspi4m0="raspi-4m-${i}"
    let i++
    local raspi4m1="raspi-4m-${i}"
    let i++
    local raspi4m2="raspi-4m-${i}"
    let i++
    local raspi4m3="raspi-4m-${i}"
    let i=clusterId*3
    local baseStation0="base-station-5g-${i}"
    let i++
    local baseStation1="base-station-5g-${i}"
    let i++
    local baseStation2="base-station-5g-${i}"

    subclusterYaml=$(cat <<EOF
# Subcluster ${clusterId}

apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${cloudNode}-to-${raspi4m0}
spec:
  nodeA: ${cloudNode}
  nodeB: ${raspi4m0}
  qos:
    qualityClass: QC20Mbps
    throughput:
      bandwidthKbps: 20000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 20
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${cloudNode}-to-${raspi4m1}
spec:
  nodeA: ${cloudNode}
  nodeB: ${raspi4m1}
  qos:
    qualityClass: QC5Mbps
    throughput:
      bandwidthKbps: 5000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 50
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${cloudNode}-to-${raspi4m2}
spec:
  nodeA: ${cloudNode}
  nodeB: ${raspi4m2}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 80
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi4m0}-to-${raspi4m1}
spec:
  nodeA: ${raspi4m0}
  nodeB: ${raspi4m1}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 80
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi4m1}-to-${raspi4m2}
spec:
  nodeA: ${raspi4m1}
  nodeB: ${raspi4m2}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 80
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b0}-to-${raspi4m1}
spec:
  nodeA: ${raspi3b0}
  nodeB: ${raspi4m1}
  qos:
    qualityClass: QC20Mbps
    throughput:
      bandwidthKbps: 20000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 5
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b1}-to-${raspi4m1}
spec:
  nodeA: ${raspi3b1}
  nodeB: ${raspi4m1}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b0}-to-${raspi4m2}
spec:
  nodeA: ${raspi3b0}
  nodeB: ${raspi4m2}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b1}-to-${raspi4m2}
spec:
  nodeA: ${raspi3b1}
  nodeB: ${raspi4m2}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b0}-to-${raspi4m3}
spec:
  nodeA: ${raspi3b0}
  nodeB: ${raspi4m3}
  qos:
    qualityClass: QC20Mbps
    throughput:
      bandwidthKbps: 20000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 20
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b0}-to-${raspi4s0}
spec:
  nodeA: ${raspi3b0}
  nodeB: ${raspi4s0}
  qos:
    qualityClass: QC20Mbps
    throughput:
      bandwidthKbps: 20000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 40
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${raspi3b1}-to-raspi-4s1
spec:
  nodeA: ${raspi3b1}
  nodeB: raspi-4s1
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 5
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation0}-to-${raspi4m3}
spec:
  nodeA: ${baseStation0}
  nodeB: ${raspi4m3}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 17000000
    latency:
      packetDelayMsec: 5
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation0}-to-${raspi4s0}
spec:
  nodeA: ${baseStation0}
  nodeB: ${raspi4s0}
  qos:
    qualityClass: QC2Mbps
    throughput:
      bandwidthKbps: 2000
      bandwidthVariance: 64000
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation1}-to-${raspi4s0}
spec:
  nodeA: ${baseStation1}
  nodeB: ${raspi4s0}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 5
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation1}-to-${raspi4s1}
spec:
  nodeA: ${baseStation1}
  nodeB: ${raspi4s1}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation2}-to-${raspi4s0}
spec:
  nodeA: ${baseStation2}
  nodeB: ${raspi4s0}
  qos:
    qualityClass: QC10Mbps
    throughput:
      bandwidthKbps: 10000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
apiVersion: cluster.k8s.rainbow-h2020.eu/v1
kind: NetworkLink
metadata:
  name: ${baseStation2}-to-${raspi4s1}
spec:
  nodeA: ${baseStation2}
  nodeB: ${raspi4s1}
  qos:
    qualityClass: QC2Mbps
    throughput:
      bandwidthKbps: 2000
      bandwidthVariance: 0
    latency:
      packetDelayMsec: 10
      packetDelayVariance: 0
    packetLoss:
      packetLossBp: 0
---
EOF
)
    RET=$(echo -e "${subclusterYaml}\n${intraCloudYaml}")
}

if [ "$1" == "" ]; then
    printUsage
    exit 1
fi

generateSubcluster $1
echo "${RET}"
