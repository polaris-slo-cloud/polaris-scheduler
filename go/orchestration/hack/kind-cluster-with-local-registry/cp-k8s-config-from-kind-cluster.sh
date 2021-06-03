#!/bin/bash

echo "Copying Kubernetes config from kind cluster to bin/config directory."
SCRIPT_DIR="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
mkdir -p $SCRIPT_DIR/../../bin/config
docker cp kind-control-plane:/etc/kubernetes $SCRIPT_DIR/../../bin/config

echo "Open bin/config/kubernetes/scheduler.conf and replace the server's address with the one of the kind-kind cluster in your KUBECONFIG file"
