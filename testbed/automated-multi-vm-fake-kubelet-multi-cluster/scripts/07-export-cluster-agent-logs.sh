#!/bin/bash
# set -x
set -o errexit

# This script exports the cluster-agent logs.
# Parameters:
# $1 - the absolute path to the results directory.
# $2 - the name of this experiment iteration. This will be used to postfix the log file name.

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

POLARIS_CLUSTER_AGENT_NAME="polaris-cluster-agent"

RESULTS_DIR="$1"
ITERATION_NAME="$2"


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "This script exports the cluster-agent logs"
    echo "Usage:"
    echo "./07-export-cluster-agent-logs.sh <result directory (absolute path, will be created)> <iteration name>"
    echo "Example:"
    echo "./07-export-cluster-agent-logs.sh \"$(pwd)/results\" \"iteration-01\""
}

function validateResultsDir() {
    if [ "$RESULTS_DIR" == "" ]; then
        printError "Please specify the absolute path of the directory (which will be created) for storing the results as the first argument."
        printUsage
        exit 1
    fi
}

function validateIterationName() {
    if [ "$ITERATION_NAME" == "" ]; then
        printError "Please specify the experiment iteration name as the second argument."
        printUsage
        exit 1
    fi
}

###############################################################################
# Script Start
###############################################################################

validateResultsDir
validateIterationName

mkdir -p "$RESULTS_DIR"
logFile="$RESULTS_DIR/cluster-agent-$ITERATION_NAME.log"

clusterAgentPod=$(kubectl get pods -n polaris -o=custom-columns="name:.metadata.name" | grep "$POLARIS_CLUSTER_AGENT_NAME")

echo "Exporting cluster-agent logs to $logFile"
kubectl logs -n polaris "$clusterAgentPod" > "$logFile"

