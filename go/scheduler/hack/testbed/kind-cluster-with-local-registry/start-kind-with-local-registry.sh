#!/bin/bash
# set -x
set -o errexit

# This script is based on https://kind.sigs.k8s.io/docs/user/local-registry/
# ToDo: Integrate this into the k8s-test-cluster setup script and docker-compose.yaml

function printUsage() {
    echo "Usage:"
    echo "./start-kind-with-local-registry.sh [nodesCount]"
    echo "Example with 10 nodes:"
    echo "./start-kind-with-local-registry.sh 10"
}

###############################################################################
# Global variables
###############################################################################

# Local Docker registry info
reg_name="kind-registry"
reg_port="5000"

# kind Kubernetes node image
kindImage="kindest/node:v1.22.9@sha256:ad5b8404c4052781365a4e70bb7d17c5331e4177bd4a7cd214339316cd6193b6"

declare -i nodesCount=1

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/common.sh"


###############################################################################
# Functions
###############################################################################

# Starts a local Docker registry unless it already exists.
function startLocalRegistry() {
    # create registry container unless it already exists
    local running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
    if [[ "${running}" != 'true' ]]; then
    docker run \
        -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
        registry:2
    fi
}

# Generates a kind cluster config.
function generateKindClusterConfig() {
    local config=$(cat <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
nodes:
- role: control-plane
  image: ${kindImage}
EOF
    )

    local singleWorkerNode=$(cat <<EOF
- role: worker
  image: ${kindImage}
EOF
    )

    local workerNodesCount=$(($nodesCount - 1))
    for i in $(seq 1 $workerNodesCount); do
        config=$(echo -e "${config}\n${singleWorkerNode}")
    done

    RET=${config}
}

# Starts a local kind cluster with the number of nodes specified in the global $nodesCount variable.
function startLocalCluster() {
    generateKindClusterConfig
    local clusterConfig=${RET}

    # create a cluster with the local registry enabled in containerd
    echo "${clusterConfig}" | kind create cluster --config=-

    # connect the registry to the cluster network
    # (the network may already be connected)
    docker network connect "kind" "${reg_name}" || true

    # Document the local registry
    # https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

    if (( $nodesCount > 1)); then
        # Remove the taint from the control plane node.
        kubectl taint --context $CONTEXT node kind-control-plane node-role.kubernetes.io/master-
    fi
}


###############################################################################
# Script Start
###############################################################################

if [[ $1 =~ ^[0-9]+$ ]]; then
    nodesCount=$1
else
    echo "Please specify the number of cluster nodes as the first argument."
    printUsage
    exit 1
fi

startLocalRegistry
startLocalCluster
