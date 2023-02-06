#!/bin/bash
# set -x
set -o errexit
set -m

# This script starts a cluster with fake-kubelet nodes and waits until all nodes are running.
# Parameters:
# $1 - the absolute path to a cluster config file.
# $2 - the absolute path to a directory containing the cluster-agent deployment YAML files.
#
# A pod that should be schedulable on one of the fake nodes needs to have the following annotation toleration:
# tolerations:
#   - key: "fake-kubelet/provider"
#     operator: "Exists"
#     effect: "NoSchedule"

###############################################################################
# Global variables and imports
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

START_SINGLE_CLUSTER_SCRIPT="${SCRIPT_DIR}/../../fake-kubelet-cluster/start-kind-with-fake-kubelet.sh"

CLUSTER_CONFIG="$1"
CLUSTER_AGENT_DEPLOYMENT_DIR="$2"

# Used by checkAllNodesAndPodsReady
LAST_READINESS_CHECK_RESULT=1


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Usage:"
    echo "./scripts/02-start-cluster.sh <cluster-config-file (absolute path)> <cluster-agent-deployment-yaml-dir (absolute path)>"
    echo "Example: "
    echo "./scripts/02-start-cluster.sh \"$(pwd)/clusters/20k-nodes/cluster-01.config.sh\" \"$(pwd)/polaris-cluster-agent/default-2-smart-sampling\""
}

function validateIterationConfigOrExit() {
    if [ "$CLUSTER_CONFIG" == "" ] || [ ! -f "$CLUSTER_CONFIG" ]; then
        printError "Please specify the absolute path of a cluster config file as the first argument."
        printUsage
        exit 1
    fi
}

function validateClusterAgentDirOrExit() {
    if [ "$CLUSTER_AGENT_DEPLOYMENT_DIR" == "" ] || [ ! -d "$CLUSTER_AGENT_DEPLOYMENT_DIR" ]; then
        printError "Please specify the absolute path of the directory with the polaris-cluster-agent deployment files as the second argument."
        printUsage
        exit 1
    fi
}

# Starts the cluster by deploying the fake nodes and the polaris-cluster-agent.
function startCluster() {
    local configPath="$1"
    local clusterAgentDir="$2"

    # Start the cluster.
    ("${START_SINGLE_CLUSTER_SCRIPT}" "$configPath")

    # Deploy the cluster agent.
    kubectl --context $CONTEXT apply -f "$clusterAgentDir"

    # Create the test namespace
    kubectl --context $CONTEXT create namespace test
}

# Checks if all nodes and pods are ready and running.
# Since new nodes are added by the fake-kubelet controllers, we want to ensure that this check succeeds twice in a row.
function checkAllNodesAndPodsReady() {
    # Check for non-ready nodes
    local notReadyNodes=$(kubectl get nodes | grep NotReady | wc -l)
    if (( $notReadyNodes == 0)); then
        RET=0
    else
        echo "$notReadyNodes nodes are not ready yet."
        LAST_READINESS_CHECK_RESULT=1
        RET=1
        return
    fi

    # Check for pods that are not running
    # Note that the header row of the kubectl output will always match the grep search expression.
    local notRunningPods=$(kubectl get pods -A | grep --invert-match "\bRunning\b" | wc -l)
    notRunningPods=$(( notRunningPods - 1 ))
    if (( $notRunningPods == 0 )); then
        RET=0
    else
        echo "$notRunningPods pods are not running yet."
        LAST_READINESS_CHECK_RESULT=1
        RET=1
        return
    fi

    # Check if the polaris-cluster-agent is running
    local polarisClusterAgentPods=$(kubectl get pods -n polaris | grep "\bRunning\b" | wc -l)
    if (( $polarisClusterAgentPods == 1 )); then
        RET=0
    else
        echo "polaris-cluster-agent pod is not running yet."
        LAST_READINESS_CHECK_RESULT=1
        RET=1
        return
    fi

    # We get here, if all checks were successful.
    # However, if this is only the first time that all checks were successful, we set RET=1, because we want to check a second time.
    if (( $LAST_READINESS_CHECK_RESULT != 0 )); then
        echo "First readiness check successful. Sleeping and checking a second time."
        RET=1
    fi
    LAST_READINESS_CHECK_RESULT=0
}


###############################################################################
# Script Start
###############################################################################

validateIterationConfigOrExit
validateClusterAgentDirOrExit
source "${CLUSTER_CONFIG}"

startCluster "$CLUSTER_CONFIG" "$CLUSTER_AGENT_DEPLOYMENT_DIR"

echo "sleep $LONG_SLEEP"
sleep $LONG_SLEEP
waitUntilDone "checkAllNodesAndPodsReady" "$LONG_SLEEP"

echo "Successfully configured the testbed components in the cluster."

