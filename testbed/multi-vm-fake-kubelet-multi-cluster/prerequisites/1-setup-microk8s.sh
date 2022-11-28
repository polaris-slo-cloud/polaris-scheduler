#!/bin/bash
# set -x
set -o errexit
set -m

# This script must be run with administrator privileges.

# Install MicroK8s
snap install microk8s --classic

# Wait for MicroK8s to be up and running
microk8s status --wait-ready

# Enable the addons that we need for our experiments.
microk8s enable dns ingress rbac
