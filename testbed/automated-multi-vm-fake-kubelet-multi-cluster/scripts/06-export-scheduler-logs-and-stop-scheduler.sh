#!/bin/bash
# set -x
set -o errexit

# This script exports the scheduler logs and then shuts down the scheduler.
# Parameters:
# $1 - the absolute path to a directory containing a polaris-scheduler docker compose file.
# $2 - the absolute path to the results directory.
# $3 - the name of this experiment iteration. This will be used to postfix the log file name.

###############################################################################
# Global variables
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

POLARIS_SCHEDULER_CONTAINER="polaris-scheduler"

SCHEDULER_COMPOSE_DIR="$1"
RESULTS_DIR="$2"
ITERATION_NAME="$3"


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "This script exports the scheduler logs and then shuts down the scheduler"
    echo "Usage:"
    echo "./06-export-scheduler-logs-and-stop-scheduler.sh <polaris-scheduler docker-compose directory (absolute path)> <result directory (absolute path, will be created)> <iteration name>"
    echo "Example:"
    echo "./06-export-scheduler-logs-and-stop-scheduler.sh \"\$(pwd)/polaris-scheduler/default-config\" \"$(pwd)/results\" \"iteration-01\""
}

function validateSchedulerComposeDirOrExit() {
    if [ "$SCHEDULER_COMPOSE_DIR" == "" ] || [ ! -d "$SCHEDULER_COMPOSE_DIR" ]; then
        printError "Please specify the absolute path of the directory with the polaris-scheduler docker-compose.yml file as the first argument."
        printUsage
        exit 1
    fi
}

function validateResultsDir() {
    if [ "$RESULTS_DIR" == "" ]; then
        printError "Please specify the absolute path of the directory (which will be created) for storing the results as the second argument."
        printUsage
        exit 1
    fi
}

function validateIterationName() {
    if [ "$ITERATION_NAME" == "" ]; then
        printError "Please specify the experiment iteration name as the third argument."
        printUsage
        exit 1
    fi
}

###############################################################################
# Script Start
###############################################################################

validateSchedulerComposeDirOrExit
validateResultsDir
validateIterationName

cd "$SCHEDULER_COMPOSE_DIR"

mkdir -p "$RESULTS_DIR"
logFile="$RESULTS_DIR/scheduler-$ITERATION_NAME.log"

echo "Exporting scheduler logs to $logFile"
docker compose logs "$POLARIS_SCHEDULER_CONTAINER" > "$logFile"

echo "Shutting down scheduler"
docker compose down
