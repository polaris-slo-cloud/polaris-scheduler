#!/bin/bash

# kind Kubernetes node image
kindImage="kindest/node:v1.24.3@sha256:e812632818ae40cee2a9a22123ceb075f8550d132f860d67e2c35b7caf0fa215"

# fake-kubelet image
fakeKubeletImageVersionTag="v0.8.0"

# Declares the types of fake nodes and how many nodes of each type to create.
# For each fake node type, the amount of CPUs and memory must be added to fakeNodeTypeCpus and fakeNodeTypeMemory respectively.
declare -A fakeNodeTypes=(
    ["raspi-3b-plus"]="2"
    ["raspi-4b-2gi"]="2"
    ["raspi-4b-4gi"]="4"
    ["cell-5g-base-station"]="3"
    ["cloud-medium"]="1"
)

# Each node's CPUs are configured as `cpu` and `polaris-slo-cloud.github.io/fake-cpu`.
declare -A fakeNodeTypeCpus=(
    ["raspi-3b-plus"]="4000m"
    ["raspi-4b-2gi"]="4000m"
    ["raspi-4b-4gi"]="4000m"
    ["cell-5g-base-station"]="4000m"
    ["cloud-medium"]="8000m"
)

# Each node's memory is configured as `memory` and `polaris-slo-cloud.github.io/fake-memory`.
declare -A fakeNodeTypeMemory=(
    ["raspi-3b-plus"]="1Gi"
    ["raspi-4b-2gi"]="2Gi"
    ["raspi-4b-4gi"]="4Gi"
    ["cell-5g-base-station"]="1Gi"
    ["cloud-medium"]="16Gi"
)

# Optional extra node labels for each node type.
# The value for each node type has to be a string of the following format (slashes and quotes must be escaped):
# "<domain1.io>\/<label1>: <value1>;<domain2.io>\/<label2>: <value2>;<...>"
declare -A extraNodeLabels=(
    ["cell-5g-base-station"]="polaris-slo-cloud.github.io\/base-station-5g: \"\";polaris-slo-cloud.github.io\/test-label: \"true\""
)

# Extended resources.
# The value for each node type has to be a string of the following format (slashes must be escaped):
# "<domain1.io>\/<resource1>: <count1>;<domain2.io>\/<resource2>: <count2>;<...>"
declare -A extendedResources=(
    ["cell-5g-base-station"]="polaris-slo-cloud.github.io\/base-station-5g: 1;polaris-slo-cloud.github.io\/test-resource: 1"
)
