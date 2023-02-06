#!/bin/bash
# set -x
set -o errexit
set -m

# This script starts the polaris-scheduler and waits until it has completed initialization.
# Parameters:
# $1 - the absolute path to a directory containing a polaris-scheduler docker compose file.

###############################################################################
# Global variables and imports
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

# This script is used to check if the scheduler has completed initialization.
# It waits until the log output does not change any more.
WAIT_UNTIL_LOGS_DO_NOT_CHANGE_SH="${SCRIPT_DIR}/05-wait-until-experiment-done.sh"

SCHEDULER_COMPOSE_DIR="$1"


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "Usage:"
    echo "./scripts/03-start-polaris-scheduler.sh <polaris-scheduler docker-compose directory (absolute path)>"
    echo "Example: "
    echo "./scripts/03-start-polaris-scheduler.sh \"\$(pwd)/polaris-scheduler/default-config\""
}

function validateSchedulerComposeDirOrExit() {
    if [ "$SCHEDULER_COMPOSE_DIR" == "" ] || [ ! -d "$SCHEDULER_COMPOSE_DIR" ]; then
        printError "Please specify the absolute path of the directory with the polaris-scheduler docker-compose.yml file as the first argument."
        printUsage
        exit 1
    fi
}


###############################################################################
# Script Start
###############################################################################

validateSchedulerComposeDirOrExit

cd "$SCHEDULER_COMPOSE_DIR"
docker compose up -d

("$WAIT_UNTIL_LOGS_DO_NOT_CHANGE_SH" "$SCHEDULER_COMPOSE_DIR")

echo "Successfully started polaris-scheduler."

