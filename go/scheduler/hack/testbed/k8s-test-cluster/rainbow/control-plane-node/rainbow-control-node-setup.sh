#!/bin/sh
# This script sets up a Kubernetes control plane node.

# This file marks that the node has already been set up.
SETUP_COMPLETE_FILE=/rainbow/node-setup-complete
if [ -f $SETUP_COMPLETE_FILE ]
then
    echo "This node has already been set up. Exiting setup script."
    exit 0
fi

# The directory with the setup files for the control plane node
CONTROL_NODE_DIR=/rainbow/control-plane-node

# The location, where we want to place the KUBECONFIG file for other nodes
KUBECONFIG_DEST=/rainbow/kubeconfig/config

# The KUBECONFIG file that should be usable on the host machine.
KUBECONFIG_DEST_PUBLIC=${KUBECONFIG_DEST}.public

# Delete the KUBECONFIG files if they already exist
rm -f $KUBECONFIG_DEST
rm -f $KUBECONFIG_DEST_PUBLIC

# Wait for systemd to be running
while [ ! $(pidof "/sbin/init") ]; do
    echo "Waiting for /sbin/init to start"
    sleep 1
done

echo "Setting up Kubernetes control plane node"

# Turn on DEBUG output
set -x
# Exit immediately if a command fails
set -o errexit
# Do not allow the use of unset environment variables
set -o nounset

# Get the IP address of the node.
IP_ADDR=$(ip a s eth0 | egrep -o 'inet [0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | cut -d' ' -f2)
echo "IP Adress: ${IP_ADDR}"

# Write the IP Address to kubeadm.conf
sed -e "s/{{ \.NodeIp }}/${IP_ADDR}/" ${CONTROL_NODE_DIR}/kubeadm.conf.template > ${CONTROL_NODE_DIR}/kubeadm.conf
cat ${CONTROL_NODE_DIR}/kubeadm.conf

# Initialize the Kubernetes cluster
kubeadm init --skip-phases=preflight --config=${CONTROL_NODE_DIR}/kubeadm.conf --skip-token-print --v=6

# Install the CNI
kubectl create --kubeconfig=/etc/kubernetes/admin.conf -f ${CONTROL_NODE_DIR}/manifests/default-cni.yaml

# Install a StorageClass
kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f ${CONTROL_NODE_DIR}/manifests/default-storage.yaml

# Copy the kubeconfig file to make it available to kubectl within the container for debugging purposes.
mkdir -p $HOME/.kube
cp /etc/kubernetes/admin.conf $HOME/.kube/config

# Copy the kubeconfig file to make it available outside the container.
# The existence of this file will trigger the setup scripts in the worker nodes.
cp /etc/kubernetes/admin.conf $KUBECONFIG_DEST
sed -e "s/control-plane\:6443/localhost:36443\n    tls-server-name: control-plane/" $KUBECONFIG_DEST > $KUBECONFIG_DEST_PUBLIC
chmod a+rw $KUBECONFIG_DEST
chmod a+rw $KUBECONFIG_DEST_PUBLIC

touch $SETUP_COMPLETE_FILE
echo "Control-plane node setup complete"
