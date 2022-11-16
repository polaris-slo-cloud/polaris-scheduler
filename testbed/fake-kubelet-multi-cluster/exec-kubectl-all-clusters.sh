#!/bin/bash
# set -x
set -o errexit
set -m

# Executes kubectl with the specified arguments on all clusters created with "start-kind-multi-cluster.sh"

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/config.sh"


###############################################################################
# Functions
###############################################################################

function execKubectl() {
    local configPath="$1"
    local cmd="$2"
    (
        source "$configPath"
        CONTEXT="kind-${kindClusterName}"
        echo "kubectl --context $CONTEXT $cmd"
        kubectl --context $CONTEXT $cmd
    )
}

function printUsage() {
    echo "Execute kubectl with the specified arguments on all clusters configured in config.sh"
    echo "Usage: ./exec-kubectl-all-clusters.sh [kubectl args]"
    echo "Example: ./exec-kubectl-all-clusters.sh create namespace test"
}


###############################################################################
# Script Start
###############################################################################

kubectlCmd="$@"

if [ -z "$kubectlCmd" ]; then
    printUsage
    exit 0
fi

for config in "${clusterConfigs[@]}"; do
    execKubectl "${SCRIPT_DIR}/${config}" "$kubectlCmd"
done

echo "Successfully executed kubectl on ${#clusterConfigs[@]} clusters."
