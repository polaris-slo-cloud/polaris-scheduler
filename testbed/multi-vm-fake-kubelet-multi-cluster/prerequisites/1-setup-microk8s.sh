#!/bin/bash
# set -x
set -o errexit
set -m

# To change the Kubernetes version, either set the MICRO_K8S_CHANNEL environment variable or modify the assignment below.
if [ "$MICRO_K8S_CHANNEL" == "" ]; then
    MICRO_K8S_CHANNEL="1.25/stable"
fi

# This script must be run with administrator privileges.
if [ "$SUDO_USER" == "" ]; then
    echo "This script must be run with sudo."
    exit 1
fi
userHome=$(eval echo ~$SUDO_USER)

# Install MicroK8s
snap install microk8s --channel=$MICRO_K8S_CHANNEL --classic

# Wait for MicroK8s to be up and running
microk8s status --wait-ready

# Enable the addons that we need for our experiments.
microk8s enable dns ingress rbac

# Install kubectl and export the MicroK8s kubeconfig
snap install kubectl --channel=$MICRO_K8S_CHANNEL --classic
mkdir -p "${userHome}/.kube"
microk8s config > "${userHome}/.kube/config"
kubectl completion bash | tee /etc/bash_completion.d/kubectl > /dev/null

# Add the current user to the microk8s group.
usermod -a -G microk8s $SUDO_USER
chown -f -R $SUDO_USER "${userHome}/.kube"

kubectlContext=$(microk8s kubectl config current-context)
echo "MicroK8s setup complete."
echo "In the experiment scripts, please ensure that you have set the kubectl context to '$kubectlContext'"
echo "Please run 'newgrp microk8s' or reboot the system to reload the current user's groups."
