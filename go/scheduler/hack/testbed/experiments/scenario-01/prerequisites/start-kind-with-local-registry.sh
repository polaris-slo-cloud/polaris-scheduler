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

# Deploy our schedulers
# kubectl apply -f ./scheduler-deployments/comparison-scheduler.yaml
# kubectl apply -f ./scheduler-deployments/rainbow-scheduler.yaml

echo "Please run ./create-extended-resources.sh with kubectl proxy"
