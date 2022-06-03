#!/bin/bash
# set -x
set -o errexit

# This script starts a single node kind cluster and deploys fake-kubelet (https://github.com/wzshiming/fake-kubelet) to create simulated nodes.

###############################################################################
# Global variables
###############################################################################

# kind Kubernetes node image
kindImage="kindest/node:v1.22.9@sha256:ad5b8404c4052781365a4e70bb7d17c5331e4177bd4a7cd214339316cd6193b6"

declare -i nodesCount=1

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/common.sh"


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Usage:"
    echo "./start-kind-with-fake-kubelet.sh [nodesCount]"
    echo "Example with 10 nodes:"
    echo "./start-kind-with-fake-kubelet.sh 10"
}

# Starts a local kind cluster with a single node.
function startLocalCluster() {
    kind create cluster --image ${kindImage}

    # Ensure that we do not schedule anything on the control plane node.
    kubectl taint --context $CONTEXT node kind-control-plane node-role.kubernetes.io/master=:NoSchedule
}

# Deploys fake-kubelet to simulate nodes.
function deployFakeKubelet() {
    kubectl --context $CONTEXT apply -f "${SCRIPT_DIR}/fake-kubelet/"
}


###############################################################################
# Script Start
###############################################################################

# if [[ $1 =~ ^[0-9]+$ ]]; then
#     nodesCount=$1
# else
#     echo "Please specify the number of cluster nodes as the first argument."
#     printUsage
#     exit 1
# fi

startLocalCluster
deployFakeKubelet
