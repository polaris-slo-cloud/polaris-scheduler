#!/bin/bash
# set -x
set -o errexit
set -m

# This script starts multiple kind clusters with fake-kubelet nodes.
# A pod that should be schedulable on one of the fake nodes needs to have the following annotation toleration:
# tolerations:
#   - key: "fake-kubelet/provider"
#     operator: "Exists"
#     effect: "NoSchedule"

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/config.sh"

START_SINGLE_CLUSTER_SCRIPT="${SCRIPT_DIR}/../fake-kubelet-cluster/start-kind-with-fake-kubelet.sh"
CLUSTER_AGENT_DEPLOYMENT_YAML="${SCRIPT_DIR}/../../go/cluster-agent/manifests/deployment-k8s"
CLUSTER_AGENT_SERVICE_YAML="${SCRIPT_DIR}/deployments/polaris-cluster-agent-service.yaml"

CLUSTER_AGENT_NODE_PORT=30033


###############################################################################
# Functions
###############################################################################

function startCluster() {
    local configPath="$1"

    # Start the cluster.
    ("${START_SINGLE_CLUSTER_SCRIPT}" "$configPath")

    # Deploy the cluster agent.
    (
        source "$configPath"
        # Set the kubectl context according to the cluster name.
        CONTEXT="kind-${kindClusterName}"

        kubectl --context $CONTEXT apply -f "$CLUSTER_AGENT_DEPLOYMENT_YAML"
        kubectl --context $CONTEXT apply -f "$CLUSTER_AGENT_SERVICE_YAML"
    )
}


###############################################################################
# Script Start
###############################################################################

for config in "${clusterConfigs[@]}"; do
    startCluster "${SCRIPT_DIR}/${config}"
done

echo "Successfully created ${#clusterConfigs[@]} clusters."
