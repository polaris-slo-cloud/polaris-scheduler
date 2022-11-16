#!/bin/bash
# set -x
set -o errexit
set -m

# Stops the kind multi-cluster created with "start-kind-multi-cluster.sh"

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/config.sh"


###############################################################################
# Functions
###############################################################################

function stopCluster() {
    local configPath="$1"
    (
        source "$configPath"
        kind delete cluster --name $kindClusterName
    )
}


###############################################################################
# Script Start
###############################################################################

for config in "${clusterConfigs[@]}"; do
    stopCluster "${SCRIPT_DIR}/${config}"
done

echo "Successfully stopped ${#clusterConfigs[@]} clusters."
