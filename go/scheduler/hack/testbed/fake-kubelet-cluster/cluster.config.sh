#!/bin/bash

# kind Kubernetes node image
kindImage="kindest/node:v1.22.9@sha256:ad5b8404c4052781365a4e70bb7d17c5331e4177bd4a7cd214339316cd6193b6"

# Declares the types of fake nodes and how many nodes of each type to create.
# For each fake node type, the amount of CPUs and memory must be added to fakeNodeTypeCpus and fakeNodeTypeMemory respectively.
declare -A fakeNodeTypes=(
    ["raspi-3b-plus"]="5"
    ["raspi-4b-2gi"]="2"
    ["raspi-4b-4gi"]="4"
    ["cloud-medium"]="1"
)

declare -A fakeNodeTypeCpus=(
    ["raspi-3b-plus"]="4000m"
    ["raspi-4b-2gi"]="4000m"
    ["raspi-4b-4gi"]="4000m"
    ["cloud-medium"]="8000m"
)

declare -A fakeNodeTypeMemory=(
    ["raspi-3b-plus"]="1Gi"
    ["raspi-4b-2gi"]="2Gi"
    ["raspi-4b-4gi"]="4Gi"
    ["cloud-medium"]="16Gi"
)

# Additional extended resources.
# The keys are composed of the node name and the name of the resource:
# "<node-name>:<resource-name>"
declare -A customResources=(
    ["kind-worker:rainbow-h2020.eu~1camera"]="1"
    ["kind-worker2:rainbow-h2020.eu~1camera"]="1"
    ["kind-worker3:rainbow-h2020.eu~1camera"]="1"
)
