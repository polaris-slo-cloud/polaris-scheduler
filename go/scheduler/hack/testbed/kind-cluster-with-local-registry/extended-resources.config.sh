#!/bin/bash

# The number of nodes in the cluster.
# This determines how many slots of the arrays below will actually be read and applied to the cluster's nodes.
NODES_COUNT=10

# Name of the fake CPU resource.
FAKE_CPU_RESOURCE_NAME="rainbow-h2020.eu~1fake-cpu"

# Name of the fake memory resource.
FAKE_MEMORY_RESOURCE_NAME="rainbow-h2020.eu~1fake-memory"

# Name of the node cost label.
NODE_COST_LABEL_NAME="rainbow-h2020.eu~1node-cost-per-hour"

# Fake CPUs in millicores
declare -A fakeCpus=(
    # Fog nodes
    ["kind-control-plane"]="4000"
    ["kind-worker"]="4000"
    ["kind-worker2"]="4000"
    ["kind-worker3"]="4000"
    ["kind-worker4"]="4000"
    ["kind-worker5"]="4000"
    ["kind-worker6"]="4000"

    # Cloud nodes (represent node types, the configured resources are actualAvailableResource * 1000)
    ["kind-worker7"]="4000000"
    ["kind-worker8"]="8000000"
    ["kind-worker9"]="16000000"
)

# Fake memory in MiB
declare -A fakeMemory=(
    # Fog nodes
    ["kind-control-plane"]="2048"
    ["kind-worker"]="1024"
    ["kind-worker2"]="1024"
    ["kind-worker3"]="2048"
    ["kind-worker4"]="1024"
    ["kind-worker5"]="1024"
    ["kind-worker6"]="2048"

    # Cloud nodes (represent node types, the configured resources are actualAvailableResource * 1000)
    ["kind-worker7"]="2048000"
    ["kind-worker8"]="8192000"
    ["kind-worker9"]="16384000"
)

# Costs of the nodes.
declare -A nodeCosts=(
    # Fog nodes
    ["kind-control-plane"]="2.00"
    ["kind-worker"]="1.50"
    ["kind-worker2"]="1.50"
    ["kind-worker3"]="2.00"
    ["kind-worker4"]="1.50"
    ["kind-worker5"]="1.50"
    ["kind-worker6"]="2.00"

    # Cloud nodes
    ["kind-worker7"]="2.00"
    ["kind-worker8"]="3.00"
    ["kind-worker9"]="4.00"
)
