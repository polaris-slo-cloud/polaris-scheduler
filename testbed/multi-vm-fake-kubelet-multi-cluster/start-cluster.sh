#!/bin/bash
# set -x
set -o errexit
set -m

# This script starts the kind cluster with fake-kubelet nodes, as specified in the config file passed as a parameter.
# A pod that should be schedulable on one of the fake nodes needs to have the following annotation toleration:
# tolerations:
#   - key: "fake-kubelet/provider"
#     operator: "Exists"
#     effect: "NoSchedule"

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")

START_SINGLE_CLUSTER_SCRIPT="${SCRIPT_DIR}/../fake-kubelet-cluster/start-kind-with-fake-kubelet.sh"
CLUSTER_AGENT_DEPLOYMENT_YAML="${SCRIPT_DIR}/polaris-cluster-agent"

CLUSTER_AGENT_NODE_PORT=30033


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Usage:"
    echo "./start-cluster.sh <path-to-config-file>"
    echo "Example:"
    echo "./start-cluster.sh ./clusters/cluster-01.config.sh"
    echo "As an alternative you can also set the path to the config file in the POLARIS_TESTBED_CONFIG environment variable."
    echo "Example:"
    echo "export POLARIS_TESTBED_CONFIG=./clusters/cluster-01.config.sh"
    echo "./start-cluster.sh"
}

function startCluster() {
    local configPath="$1"

    # Start the cluster.
    ("${START_SINGLE_CLUSTER_SCRIPT}" "$configPath")

    # Deploy the cluster agent.
    (
        source "$configPath"

        if [ "${skipKindClusterSetup}" != true ]; then
            # Set the kubectl context according to the cluster name.
            CONTEXT="kind-${kindClusterName}"
        fi

        kubectl --context $CONTEXT apply -f "$CLUSTER_AGENT_DEPLOYMENT_YAML"
    )
}


###############################################################################
# Script Start
###############################################################################

echo $POLARIS_TESTBED_CONFIG
if [ "$POLARIS_TESTBED_CONFIG" == "" ] || [ ! -f "$POLARIS_TESTBED_CONFIG" ]; then
    if [ "$1" == "" ] || [ ! -f "$1" ]; then
        printUsage
        exit 1
    fi
    POLARIS_TESTBED_CONFIG="$1"
fi

startCluster "${POLARIS_TESTBED_CONFIG}"

echo "Successfully configured the testbed components in the cluster."
