#!/bin/bash
# set -x
set -o errexit

if [ "$1" == "" ]; then
    echo "Please provide the hostname and port of the Kubernetes API proxy as an argument of the form <host>:<port>, e.g., localhost:8001."
    exit 1
fi

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")

../../../kind-cluster-with-local-registry/create-extended-resources.sh "$1" "${SCRIPT_DIR}/cluster-config.sh"
