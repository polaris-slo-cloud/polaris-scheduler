#!/bin/bash
# set -x
set -o errexit

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
source "${SCRIPT_DIR}/cluster-config.sh"

../../../kind-cluster-with-local-registry/start-kind-with-local-registry.sh ${NODES_COUNT}

# Make sure that the required images are available on all nodes.
APP_IMAGE="gcr.io/google-containers/pause:3.2"

docker pull $APP_IMAGE
kind load docker-image $APP_IMAGE

# Deploy the CRDs and the cluster topology.
kubectl apply -f "${SCRIPT_DIR}/../../prerequisites"
kubectl apply -f "${SCRIPT_DIR}/cluster-topology.yaml"
# Undeploy the rainbow-orchestrator, which is not needed for scheduler testing.
kubectl delete deployment -n rainbow-system rainbow-orchestrator-controller-manager

echo "Please run the following:"
echo "1. kubectl proxy (in a second terminal)"
echo "2. ./create-extended-resources.sh <addr-and-port-from-kubectl-proxy> (afterwards you can stop 'kubectl proxy')"
echo "3. kubectl apply -f ./scheduler-deployments"
