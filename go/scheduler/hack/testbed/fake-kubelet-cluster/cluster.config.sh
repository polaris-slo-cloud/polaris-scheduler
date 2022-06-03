#!/bin/bash

# kind Kubernetes node image
kindImage="kindest/node:v1.22.9@sha256:ad5b8404c4052781365a4e70bb7d17c5331e4177bd4a7cd214339316cd6193b6"

# Declares the types of fake nodes and how many nodes of each type to create.
# For each fake node type, the amount of CPUs and memory must be added to fakeNodeTypeCpus and fakeNodeTypeMemory respectively.
declare -A fakeNodeTypes=(
    ["raspi-3b-plus"]="2"
    ["raspi-4b-2gi"]="2"
    ["raspi-4b-4gi"]="4"
    ["cell-5g-base-station"]="3"
    ["cloud-medium"]="1"
)

declare -A fakeNodeTypeCpus=(
    ["raspi-3b-plus"]="4000m"
    ["raspi-4b-2gi"]="4000m"
    ["raspi-4b-4gi"]="4000m"
    ["cell-5g-base-station"]="4000m"
    ["cloud-medium"]="8000m"
)

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


# Additional extended resources.
# The keys are composed of the node name and the name of the resource:
# "<fake-node-type>:<resource-name>"
declare -A fakeNodeTypeExtendedResources=(
    ["cell-5g-base-station:polaris-slo-cloud.github.io~1base-station-5g"]="1"
)
