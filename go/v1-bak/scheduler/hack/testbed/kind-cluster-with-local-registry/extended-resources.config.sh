#!/bin/bash

# The number of nodes in the cluster.
# This determines how many slots of the arrays below will actually be read and applied to the cluster's nodes.
NODES_COUNT=7

# Name of the fake CPU resource.
FAKE_CPU_RESOURCE_NAME="rainbow-h2020.eu~1fake-cpu"

# Name of the fake memory resource.
FAKE_MEMORY_RESOURCE_NAME="rainbow-h2020.eu~1fake-memory"

# Name of the node cost label.
NODE_COST_LABEL_NAME="rainbow-h2020.eu~1node-cost-per-hour"

# Fake CPUs in millicores
declare -A fakeCpus=(
    # Fog nodes
    ["kind-control-plane"]="4000m"
    ["kind-worker"]="4000m"
    ["kind-worker2"]="4000m"
    ["kind-worker3"]="4000m"
    ["kind-worker4"]="4000m"
    ["kind-worker5"]="4000m"
    ["kind-worker6"]="4000m"

    # Cloud nodes (represent node types, the configured resources are actualAvailableResource * 1000)
    # ["kind-worker7"]="4000000m"
    # ["kind-worker8"]="8000000m"
    # ["kind-worker9"]="16000000m"
)

# Fake memory in MiB
declare -A fakeMemory=(
    # Fog nodes
    ["kind-control-plane"]="8Gi"
    ["kind-worker"]="1Gi"
    ["kind-worker2"]="1Gi"
    ["kind-worker3"]="1Gi"
    ["kind-worker4"]="4Gi"
    ["kind-worker5"]="2Gi"
    ["kind-worker6"]="2Gi"

    # Cloud nodes (represent node types, the configured resources are actualAvailableResource * 1000)
    # ["kind-worker7"]="2000Gi"
    # ["kind-worker8"]="8000Gi"
    # ["kind-worker9"]="16000Gi"
)

# Costs of the nodes.
declare -A nodeCosts=(
    # Fog nodes
    ["kind-control-plane"]="1.00"
    ["kind-worker"]="1.00"
    ["kind-worker2"]="1.00"
    ["kind-worker3"]="1.00"
    ["kind-worker4"]="1.00"
    ["kind-worker5"]="1.00"
    ["kind-worker6"]="1.00"

    # Cloud nodes
    # ["kind-worker7"]="2.00"
    # ["kind-worker8"]="3.00"
    # ["kind-worker9"]="4.00"
)

# Additional extended resources.
# The keys are composed of the node name and the name of the resource:
# "<node-name>:<resource-name>"
declare -A customResources=(
    ["kind-worker:rainbow-h2020.eu~1camera"]="1"
    ["kind-worker2:rainbow-h2020.eu~1camera"]="1"
    ["kind-worker3:rainbow-h2020.eu~1camera"]="1"
)
