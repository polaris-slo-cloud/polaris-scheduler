#!/bin/bash

# This script adds extended resources to the nodes of the cluster.
# This is necessary, because it is not easily possible to mock CPU and memory resource bounds,
# thus Huang-Wei from #sig-scheduling recommended using extended resources for this purpose:
# rainbow-h2020.eu/fake-cpu
# rainbow-h2020.eu/fake-memory

# set -x
set -o errexit

if [ "$1" == "" ]; then
    echo "Please provide the hostname and port of the Kubernetes API proxy as an argument of the form <host>:<port>, e.g., localhost:8080."
    exit
fi

FAKE_CPU_RESOURCE_NAME="rainbow-h2020.eu~1fake-cpu"
FAKE_MEMORY_RESOURCE_NAME="rainbow-h2020.eu~1fake-memory"

baseUrl=$1
nodes=(
    # Fog nodes (correspond to either a Raspberry Pi 3 Model B+ (4 CPU cores, 1GB RAM) or a Raspberry Pi 4 Model B (4 CPU cores, 2GB, 4GB, or 8GB RAM)):
    "kind-control-plane" "kind-worker" "kind-worker2" "kind-worker3" "kind-worker4" "kind-worker5" "kind-worker6"
    # Cloud nodes:
    "kind-worker7" "kind-worker8" "kind-worker9"
)


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

for i in "${!nodes[@]}"; do
    nodeName=${nodes[$i]}
    nodeCpus=${fakeCpus[$nodeName]}
    nodeMemory=${fakeMemory[$nodeName]}
    echo "node: $nodeName, cpus: $nodeCpus, memory: $nodeMemory"

    curl --header "Content-Type: application/json-patch+json" \
        --request PATCH \
        --data "[{\"op\": \"add\", \"path\": \"/status/capacity/${FAKE_CPU_RESOURCE_NAME}\", \"value\": \"${nodeCpus}\"}, {\"op\": \"add\", \"path\": \"/status/capacity/${FAKE_MEMORY_RESOURCE_NAME}\", \"value\": \"${nodeMemory}\"}]" \
        "http://${baseUrl}/api/v1/nodes/${nodeName}/status"
done
