#!/bin/bash
set -x
set -o errexit

# This is a modified version of the script from https://kind.sigs.k8s.io/docs/user/local-registry/
# ToDo: Integrate this into the k8s-test-cluster setup script and docker-compose.yaml

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# create a cluster with the local registry enabled in containerd
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
nodes:
- role: control-plane
  image: kindest/node:v1.20.2@sha256:15d3b5c4f521a84896ed1ead1b14e4774d02202d5c65ab68f30eeaf310a3b1a7
- role: worker
  image: kindest/node:v1.20.2@sha256:15d3b5c4f521a84896ed1ead1b14e4774d02202d5c65ab68f30eeaf310a3b1a7
- role: worker
  image: kindest/node:v1.20.2@sha256:15d3b5c4f521a84896ed1ead1b14e4774d02202d5c65ab68f30eeaf310a3b1a7
- role: worker
  image: kindest/node:v1.20.2@sha256:15d3b5c4f521a84896ed1ead1b14e4774d02202d5c65ab68f30eeaf310a3b1a7
EOF

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

# Add labels to the nodes
FOG_WORKER_NODES=("kind-worker" "kind-worker2" "kind-worker3") # "kind-worker4" "kind-worker5" "kind-worker6")
CLOUD_WORKER_NODES=() # "kind-worker7" "kind-worker8" "kind-worker9")
CONTEXT="kind-kind"

kubectl label --context $CONTEXT node kind-control-plane node-role.kubernetes.io/fog-region-head=""
kubectl label --context $CONTEXT node kind-control-plane node-role.kubernetes.io/fog=""

for i in ${FOG_WORKER_NODES[@]}; do
    kubectl label --context $CONTEXT node $i node-role.kubernetes.io/worker=""
    kubectl label --context $CONTEXT node $i node-role.kubernetes.io/fog=""
done

for i in ${CLOUD_WORKER_NODES[@]}; do
    kubectl label --context $CONTEXT node $i node-role.kubernetes.io/worker=""
    kubectl label --context $CONTEXT node $i node-role.kubernetes.io/cloud=""
done

# Remove the taint from the fog-region-head node.
kubectl taint --context $CONTEXT node kind-control-plane node-role.kubernetes.io/master-


# Make sure that the required images are available on all nodes
# RABBIT_MQ_IMAGE="rabbitmq:3.8-alpine"
# TAXI_IMAGE="rainbowh2020/taxi-async:0.0.1"
#
# docker pull $RABBIT_MQ_IMAGE
# docker pull $TAXI_IMAGE
# kind load docker-image $RABBIT_MQ_IMAGE
# kind load docker-image $TAXI_IMAGE
