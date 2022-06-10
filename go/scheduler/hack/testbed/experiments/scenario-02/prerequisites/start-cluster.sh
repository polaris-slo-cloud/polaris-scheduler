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

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
CLUSTER_CONFIG="${SCRIPT_DIR}/cluster.config.sh"

source "${CLUSTER_CONFIG}"
(source "${SCRIPT_DIR}/../../../fake-kubelet-cluster/start-kind-with-fake-kubelet.sh" "${CLUSTER_CONFIG}")

# Install the CRDs.
kubectl apply -f "${SCRIPT_DIR}/../../prerequisites/"

# Generate the cluster topology.
let maxClusterId=subclustersCount-1
for i in $(seq 0 $maxClusterId); do
    yaml=$(bash "${SCRIPT_DIR}/gen-cluster-topology.sh" $i)
    kubectl apply -f - <<< "$yaml"
done
