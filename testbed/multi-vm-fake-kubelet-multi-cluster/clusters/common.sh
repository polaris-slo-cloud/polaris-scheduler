#!/bin/bash

# kind Kubernetes node image (required when not using MicroK8s)
# kindImage="kindest/node:v1.25.3@sha256:3f251a73d58a0db2950d5abfa5adfa503099ac1b3811e9bc253ff03c079e108e"

# Set this to true to skip setting up a kind cluster.
# In such a case a Kubernets cluster must already be running and the $CONTEXT variable (see below) must be set.
skipKindClusterSetup=true
CONTEXT=microk8s

# fake-kubelet image
fakeKubeletImageVersionTag="v0.8.0"
