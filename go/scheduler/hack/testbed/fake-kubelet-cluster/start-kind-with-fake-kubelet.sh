#!/bin/bash
# set -x
set -o errexit
set -m

# This script starts a single node kind cluster and deploys fake-kubelet (https://github.com/wzshiming/fake-kubelet) to create simulated nodes.
# A pod that should be schedulable on one of the fake nodes needs to have the following annotation toleration:
# tolerations:
#   - key: "fake-kubelet/provider"
#     operator: "Exists"
#     effect: "NoSchedule"

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/common.sh"

API_PROXY_PORT="8001"
API_PROXY_BASE_URL="localhost:${API_PROXY_PORT}"

# Special indents for formatting raw YAML strings for the template.
EXTRA_NODE_LABELS_INDENT="        "


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Usage:"
    echo "./start-kind-with-fake-kubelet.sh <path-to-config-file>"
    echo "Example:"
    echo "./start-kind-with-fake-kubelet.sh cluster.config.sh"
}

# Prints an error and exits if the configuration is not valid.
function validateConfig() {
    if [ "${kindImage}" == "" ]; then
        echo "Error: The kindImage has not been set in the configuration!"
        exit
    fi
    for fakeNodeType in "${!fakeNodeTypes[@]}"; do
        if [ "${fakeNodeTypeCpus[$fakeNodeType]}" == "" ]; then
            echo "Error: No entry for ${fakeNodeType} found in the 'fakeNodeTypeCpus' array!"
            exit 1
        fi
        if [ "${fakeNodeTypeMemory[$fakeNodeType]}" == "" ]; then
            echo "Error: No entry for ${fakeNodeType} found in the 'fakeNodeTypeMemory' array!"
            exit 1
        fi
    done
}

# Starts a local kind cluster with a single node.
function startLocalCluster() {
    kind create cluster --image ${kindImage}

    # Ensure that we do not schedule anything on the control plane node.
    kubectl taint --context $CONTEXT node kind-control-plane node-role.kubernetes.io/master=:NoSchedule
}

# Deploys fake-kubelet to simulate nodes.
function deployFakeKubelet() {
    kubectl --context $CONTEXT apply -f "${SCRIPT_DIR}/fake-kubelet/fake-kubelet-base.yaml"

    local nodesTemplateBase=$(cat "${SCRIPT_DIR}/fake-kubelet/fake-kubelet-nodes-template.yaml")

    for fakeNodeType in "${!fakeNodeTypes[@]}"; do
        local fakeNodesCount="${fakeNodeTypes[$fakeNodeType]}"
        local fakeCpus="${fakeNodeTypeCpus[$fakeNodeType]}"
        local fakeMemory="${fakeNodeTypeMemory[$fakeNodeType]}"
        getExtraNodeLabels "${fakeNodeType}"
        local extraLabels="${RET}"

        echo "Creating ${fakeNodesCount} nodes of type ${fakeNodeType} with ${fakeCpus} CPUs and ${fakeMemory} RAM."
        if [ "${extraLabels}" != "" ]; then
            echo "Extra node labels: ${extraLabels}"
        fi

        local nodeTypeYaml=$(echo "${nodesTemplateBase}" | sed -e "s/{{ \.polarisTemplate\.fakeNodeType }}/${fakeNodeType}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeNodesCount }}/${fakeNodesCount}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeCPUs }}/${fakeCpus}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeMemory }}/${fakeMemory}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.extraNodeLabels }}/${extraLabels}/" -)
        echo "${nodeTypeYaml}" | kubectl apply -f -

    done
}

# Gets the extra node labels formatted YAML string for the current fakeNodeType
function getExtraNodeLabels() {
    local fakeNodeType=$1
    RET=""

    local rawLabels=${extraNodeLabels[$fakeNodeType]}
    if [ "${rawLabels}" == "" ]; then
        return
    fi

    readarray -d ";" -t labels <<< "${rawLabels}"
    for label in "${labels[@]}"; do
        local trimmedLabel=$(echo "${label}" | tr -d "\n")
        RET="${RET}\n${EXTRA_NODE_LABELS_INDENT}${trimmedLabel}"
    done
}

# Creates the extended resources configured in fakeNodeTypeExtendedResources.
function createExtendedResources() {
    if [ "${#fakeNodeTypeExtendedResources[@]}" != "0" ]; then
        local sleepTime="1m"
        echo "Sleeping for ${sleepTime} to allow all nodes to be registered before creating extended resources."
        sleep ${sleepTime}
    fi

    # Run kubectl proxy in the background
    kubectl proxy --port=${API_PROXY_PORT} &
    local proxyPID=$!
    echo "Executing 'kubectl proxy' in the background. PID: $proxyPID"
    sleep 5s

    # Create extended resources.
    for compoundKey in "${!fakeNodeTypeExtendedResources[@]}"; do
        readarray -d : -t keyComponents <<< "${compoundKey}"
        local fakeNodeType="${keyComponents[0]}"
        local resourceName=$(echo "${keyComponents[1]}" | tr -d "\n")
        local resourceValue="${fakeNodeTypeExtendedResources[$compoundKey]}"
        local nodesCount="${fakeNodeTypes[$fakeNodeType]}"

        if [ "${nodesCount}" == "" ]; then
            echo "Error: Unknown fake node type ${fakeNodeType}."
            exit 1
        fi
        echo "Creating extended resource for fakeNodeType: $fakeNodeType, resource: $resourceName = $resourceValue"

        local maxIndex=$(($nodesCount - 1))
        for i in $(seq 0 $maxIndex ); do
            local nodeName="${fakeNodeType}-${i}"
            curl --header "Content-Type: application/json-patch+json" \
                --request PATCH \
                --data "[{\"op\": \"add\", \"path\": \"/status/capacity/${resourceName}\", \"value\": \"${resourceValue}\"}]" \
                "http://${API_PROXY_BASE_URL}/api/v1/nodes/${nodeName}/status"
        done
    done

    echo "Stopping kubectl proxy"
    kill -SIGTERM ${proxyPID}
}


###############################################################################
# Script Start
###############################################################################

if [ "$1" == "" ] || [ ! -f "$1" ]; then
    printUsage
    exit 1
fi


# Load the configuration (yes, it is dangerous to do it this way, but this script is only used in our experiments).
# For an example config file see: cluster.config.sh
source "$1"

validateConfig
startLocalCluster
deployFakeKubelet
createExtendedResources

echo "Successfully created cluster with fake-kubelet nodes."
