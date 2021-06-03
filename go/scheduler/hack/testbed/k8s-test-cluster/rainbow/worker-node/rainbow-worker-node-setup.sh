#!/bin/sh
# This script sets up a Kubernetes worker node.

# This file marks that the node has already been set up.
SETUP_COMPLETE_FILE=/rainbow/node-setup-complete
if [ -f $SETUP_COMPLETE_FILE ]
then
    echo "This node has already been set up. Exiting setup script."
    exit 0
fi

# The directory with the setup files for the worker node
WORKER_NODE_DIR=/rainbow/worker-node

# The location of the KUBECONFIG file provided by the control-plane
KUBECONFIG_SRC=/rainbow/kubeconfig/config

# This is far form ideal, because sleep is not a synchronization primitive, but if works for now.
# Give the control-node time to delete an existing KUBECONF, if it exists.
sleep 5

# Wait for the KUBECONF file to be provided by the control-plane.
while [ ! -f $KUBECONFIG_SRC ]; do sleep 1; done

echo "Setting up Kubernetes worker node"

# Turn on DEBUG output
set -x
# Exit immediately if a command fails
set -o errexit
# Do not allow the use of unset environment variables
set -o nounset

# Get the IP address of the node.
IP_ADDR=$(ip a s eth0 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2)
echo "IP Adress: ${IP_ADDR}"

# Write the IP Address and the hostname to kubeadm.conf
sed -e "s/{{ \.NodeIp }}/${IP_ADDR}/" -e "s/{{ \.NodeName }}/${HOSTNAME}/" ${WORKER_NODE_DIR}/kubeadm.conf.template > ${WORKER_NODE_DIR}/kubeadm.conf
cat ${WORKER_NODE_DIR}/kubeadm.conf

# Join the Kubernetes cluster
kubeadm join --config ${WORKER_NODE_DIR}/kubeadm.conf --skip-phases=preflight --v=6

# Wait until the cluster is ready
# ToDo: Add a loop here
kubectl --kubeconfig=$KUBECONFIG_SRC get nodes --selector=node-role.kubernetes.io/master -o=jsonpath='{.items..status.conditions[-1:].status}'

# Copy the kubeconfig file to make it available to kubectl within the container for debugging purposes.
mkdir -p $HOME/.kube
cp $KUBECONFIG_SRC $HOME/.kube/config

touch $SETUP_COMPLETE_FILE
echo "Worker node setup complete"
