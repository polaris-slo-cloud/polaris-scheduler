#!/bin/bash
# set -x
set -o errexit
set -m

# This script installs MicroK8s and must be run with sudo.

###############################################################################
# Global variables and imports
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

# To change the Kubernetes version, either set the MICRO_K8S_CHANNEL environment variable or modify the assignment below.
if [ "$MICRO_K8S_CHANNEL" == "" ]; then
    printError "\$MICRO_K8S_CHANNEL not set"
    exit 1
fi

# This script must be run with administrator privileges.
if [ "$SUDO_USER" == "" ]; then
    printError "This script must be run with sudo."
    exit 1
fi


###############################################################################
# Functions
###############################################################################

function checkMasterNodeIsReady() {
    notReadyNodes=$(microk8s kubectl get nodes | grep NotReady | wc -l)
    if (( $notReadyNodes == 0)); then
        RET=0
    else
        RET=1
    fi
}


###############################################################################
# Script Start
###############################################################################

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

# Wait until the master node is ready.
waitUntilDone "checkMasterNodeIsReady" "$MEDIUM_SLEEP"

echo "MicroK8s setup complete."
echo "In the experiment scripts, please ensure that you have set the kubectl context to '$kubectlContext'"
echo "Please run 'newgrp microk8s' or reboot the system to reload the current user's groups."
