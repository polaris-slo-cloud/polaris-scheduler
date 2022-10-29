#!/bin/bash

# Used as return value by functions.
RET=0

# kubectl context
CONTEXT="kind-kind"

# Returns the name of the Kubernetes node with the specified numeric id >= 0
function getNodeName() {
    local id=$1
    local ret=""

    case $id in
    0)
        ret="kind-control-plane"
        ;;
    1)
        ret="kind-worker"
        ;;
    *)
        ret="kind-worker${id}"
        ;;
    esac

    RET=$ret
}
