#!/bin/bash
# set -x
set -o errexit

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

        echo "Creating ${fakeNodesCount} nodes of type ${fakeNodeType} with ${fakeCpus} CPUs and ${fakeMemory} RAM."

        local nodeTypeYaml=$(echo "${nodesTemplateBase}" | sed -e "s/{{ \.polarisTemplate\.fakeNodeType }}/${fakeNodeType}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeNodesCount }}/${fakeNodesCount}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeCPUs }}/${fakeCpus}/" -)
        nodeTypeYaml=$(echo "${nodeTypeYaml}" | sed -e "s/{{ \.polarisTemplate\.fakeMemory }}/${fakeMemory}/" -)
        echo "${nodeTypeYaml}" | kubectl apply -f -
    done
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
