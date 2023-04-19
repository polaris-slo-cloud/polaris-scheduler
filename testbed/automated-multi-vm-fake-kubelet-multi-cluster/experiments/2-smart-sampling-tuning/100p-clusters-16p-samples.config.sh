#!/bin/bash

# IMPORTANT: All paths MUST be relative to $TESTBED_PATH_IN_REPO configured in the root experiment.config.sh file.

ITERATION_CONFIG_SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$ITERATION_CONFIG_SCRIPT_DIR/base.config.sh"

# The path of the directory containing the scheduler docker-compose files and config.
SCHEDULER_DOCKER_COMPOSE_DIR="./polaris-scheduler/2-smart-sampling-tuning/100p-clusters-16p-sample"

