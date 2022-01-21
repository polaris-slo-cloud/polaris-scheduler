#!/bin/bash

# This script adds extended resources to the nodes of the cluster.
# This is necessary, because it is not easily possible to mock CPU and memory resource bounds,
# thus Huang-Wei from #sig-scheduling recommended using extended resources for this purpose:
# rainbow-h2020.eu/fake-cpu
# rainbow-h2020.eu/fake-memory

# set -x
set -o errexit

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/common.sh"

function printUsage() {
    echo "create-extended-resources.sh"
    echo "Usage: ./create-extended-resources.sh <K8sApiProxyHost>:<port> <path-to-config-file>"
    echo "Example:"
    echo "./create-extended-resources.sh localhost:8001 extended-resources.config.sh"
}

if [ "$1" == "" ]; then
    printUsage
    exit 1
fi

if [ "$2" == "" ] || [ ! -f $2 ]; then
    printUsage
    exit 1
fi

baseUrl=$1

# Load the configuration (yes, it is dangerous to do it this way, but this script is only used in our experiments).
# For an example config file see: extended-resources.config.sh
source "$2"

# Set CPUs, memory, and node cost.
for i in $(seq 0 $(($NODES_COUNT - 1)) ); do
    getNodeName $i
    nodeName=${RET}
    nodeCpus=${fakeCpus[$nodeName]}
    nodeMemory=${fakeMemory[$nodeName]}
    nodeCost=${nodeCost[$nodeName]}
    echo "node: $nodeName, cpus: $nodeCpus, memory: $nodeMemory"

    curl --header "Content-Type: application/json-patch+json" \
        --request PATCH \
        --data "[{\"op\": \"add\", \"path\": \"/status/capacity/${FAKE_CPU_RESOURCE_NAME}\", \"value\": \"${nodeCpus}\"}, {\"op\": \"add\", \"path\": \"/status/capacity/${FAKE_MEMORY_RESOURCE_NAME}\", \"value\": \"${nodeMemory}\"}]" \
        "http://${baseUrl}/api/v1/nodes/${nodeName}/status"

    kubectl label --context $CONTEXT node ${nodeName} ${nodeName}="${nodeCost}"
done

# Set custom extended resources.
for compoundKey in "${!customResources[@]}"; do
    readarray -d : -t keyComponents <<< "$compoundKey"
    nodeName="${keyComponents[0]}"
    resourceName=$(echo "${keyComponents[1]}" | tr -d "\n")
    resourceValue="${customResources[$compoundKey]}"
    echo "node: $node, resource: $resourceName = $resourceValue"

    curl --header "Content-Type: application/json-patch+json" \
        --request PATCH \
        --data "[{\"op\": \"add\", \"path\": \"/status/capacity/${resourceName}\", \"value\": \"${resourceValue}\"}]" \
        "http://${baseUrl}/api/v1/nodes/${nodeName}/status"
done
