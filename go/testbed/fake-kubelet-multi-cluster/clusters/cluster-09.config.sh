#!/bin/bash

CLUSTER_CONFIG_DIR=$(dirname "${BASH_SOURCE}")
source "${CLUSTER_CONFIG_DIR}/common.sh"

# The name of the kind cluster.
kindClusterName="kind-09-cloud-frankfurt"

# The port on localhost, where the polaris-cluster-agent of this cluster should be exposed.
clusterAgentPortLocalhost=30009

# (optional) Additional kind node config.
# For config options see https://kind.sigs.k8s.io/docs/user/configuration/
kindExtraConfig=$(cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  # port forward ${clusterAgentPortLocalhost} on the host to 30033 on the control-plane node
  extraPortMappings:
  - containerPort: 30033
    hostPort: ${clusterAgentPortLocalhost}
    # optional: set the bind address on the host
    # 0.0.0.0 is the current default
    listenAddress: "127.0.0.1"
    # optional: set the protocol to one of TCP, UDP, SCTP.
    # TCP is the default
    # protocol: TCP
EOF
)

dataCenterLocation="50.113485878139905_8.664028308860287"

# Declares the types of fake nodes and how many nodes of each type to create.
# For each fake node type, the amount of CPUs and memory must be added to fakeNodeTypeCpus and fakeNodeTypeMemory respectively.
declare -A fakeNodeTypes=(
    ["cloud-small"]="1000"
    ["cloud-medium"]="600"
    ["cloud-large"]="400"
)

# Each node's CPUs are configured as `cpu` and `polaris-slo-cloud.github.io/fake-cpu`.
declare -A fakeNodeTypeCpus=(
    ["cloud-small"]="2000m"
    ["cloud-medium"]="4000m"
    ["cloud-large"]="8000m"
)

# Each node's memory is configured as `memory` and `polaris-slo-cloud.github.io/fake-memory`.
declare -A fakeNodeTypeMemory=(
    ["cloud-small"]="4Gi"
    ["cloud-medium"]="8Gi"
    ["cloud-large"]="16Gi"
)

# Optional extra node labels for each node type.
# The value for each node type has to be a string of the following format (slashes and quotes must be escaped):
# "<domain1.io>\/<label1>: <value1>;<domain2.io>\/<label2>: <value2>;<...>"
declare -A extraNodeLabels=(
    ["cloud-small"]="polaris-slo-cloud.github.io\/geo-location: \"${dataCenterLocation}\""
    ["cloud-medium"]="polaris-slo-cloud.github.io\/geo-location: \"${dataCenterLocation}\""
    ["cloud-large"]="polaris-slo-cloud.github.io\/geo-location: \"${dataCenterLocation}\""
)

# Extended resources.
# The value for each node type has to be a string of the following format (slashes must be escaped):
# "<domain1.io>\/<resource1>: <count1>;<domain2.io>\/<resource2>: <count2>;<...>"
declare -A extendedResources=(
    # ["cell-5g-base-station"]="polaris-slo-cloud.github.io\/base-station-5g: 1;polaris-slo-cloud.github.io\/test-resource: 1"
)
