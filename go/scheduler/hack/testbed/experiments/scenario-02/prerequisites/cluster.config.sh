#!/bin/bash

# kind Kubernetes node image
kindImage="kindest/node:v1.22.9@sha256:ad5b8404c4052781365a4e70bb7d17c5331e4177bd4a7cd214339316cd6193b6"

# Declares the types of fake nodes and how many nodes of each type to create.
# For each fake node type, the amount of CPUs and memory must be added to fakeNodeTypeCpus and fakeNodeTypeMemory respectively.
declare -A fakeNodeTypes=(
    ["raspi-3b"]="20" # Raspberry Pi Model 3B+
    ["raspi-4s"]="20" # Raspberry Pi Model 4B 2GB
    ["raspi-4m"]="40" # Raspberry Pi Model 4B 4GB
    ["base-station-5g"]="30"
    ["cloud-medium"]="10"
)

declare -A fakeNodeTypeCpus=(
    ["raspi-3b"]="4000m"
    ["raspi-4s"]="4000m"
    ["raspi-4m"]="4000m"
    ["base-station-5g"]="4000m"
    ["cloud-medium"]="8000m"
)

declare -A fakeNodeTypeMemory=(
    ["raspi-3b"]="1Gi"
    ["raspi-4s"]="2Gi"
    ["raspi-4m"]="4Gi"
    ["base-station-5g"]="1Gi"
    ["cloud-medium"]="16Gi"
)

# Optional extra node labels for each node type.
# The value for each node type has to be a string of the following format (slashes and quotes must be escaped):
# "<domain1.io>\/<label1>: <value1>;<domain2.io>\/<label2>: <value2>;<...>"
declare -A extraNodeLabels=(
    ["base-station-5g"]="polaris-slo-cloud.github.io\/base-station-5g: \"\";polaris-slo-cloud.github.io\/test-label: \"true\""
)

# Extended resources.
# The value for each node type has to be a string of the following format (slashes must be escaped):
# "<domain1.io>\/<resource1>: <count1>;<domain2.io>\/<resource2>: <count2>;<...>"
declare -A extendedResources=(
    ["base-station-5g"]="polaris-slo-cloud.github.io\/base-station-5g: 1;polaris-slo-cloud.github.io\/test-resource: 1"
)
