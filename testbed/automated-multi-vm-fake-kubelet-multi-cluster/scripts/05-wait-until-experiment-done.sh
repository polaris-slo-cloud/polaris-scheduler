#!/bin/bash
# set -x
set -o errexit

# This script waits until the polaris-scheduler container has finished processing all pods,
# i.e., until the experiment is finished.
# Parameters:
# $1 - the absolute path to a directory containing a polaris-scheduler docker compose file.

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(dirname "${BASH_SOURCE}")
POLARIS_SCHEDULER_CONTAINER="polaris-scheduler"

# Used by checkExperimentDone()
prevLogLength="-1"

###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Waits until the polaris-scheduler container has finished processing all pods,"
    echo "i.e., until the experiment is finished."
    echo "Usage:"
    echo "./05-wait-until-experiment-done.sh <absolute path of the polaris-scheduler docker-compose directory> <sleep duration for busy waiting>"
    echo "Example:"
    echo "./05-wait-until-experiment-done.sh \"\$(pwd)/polaris-scheduler/default-config\" 20s"
}

function getContainerLogLength() {
    RET=$(docker compose logs $POLARIS_SCHEDULER_CONTAINER | wc -l)
}

# Checks if the experiment is done by comparing the previously recorded container log length to the current one.
# Sets $RET to 0 if the experiment is done, and to 1 otherwise.
function checkExperimentDone() {
    getContainerLogLength
    local currLogLength="$RET"
    echo "prevLogLength = $prevLogLength; currLogLength = $currLogLength"

    if (( "$prevLogLength" == "$currLogLength" )); then
        RET=0
    else
        prevLogLength="$currLogLength"
        RET=1
    fi
}

###############################################################################
# Script Start
###############################################################################

source "$SCRIPT_DIR/lib/util.sh"

if [ "$1" == "" ] || [ ! -d "$1" ]; then
    printUsage
    exit 1
fi
if [ "$2" == "" ]; then
    printUsage
    exit 1
fi

cd "$1"
waitUntilDone "checkExperimentDone" "$2"
