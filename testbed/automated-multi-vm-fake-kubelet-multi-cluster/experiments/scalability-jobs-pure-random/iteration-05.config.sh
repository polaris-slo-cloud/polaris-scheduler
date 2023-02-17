#!/bin/bash

# IMPORTANT: All paths MUST be relative to $TESTBED_PATH_IN_REPO configured in the root experiment.config.sh file.

ITERATION_CONFIG_SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$ITERATION_CONFIG_SCRIPT_DIR/base.config.sh"

# Path of the JMeter test plan file.
JMETER_TEST_PLAN="./jmeter-test-plans/heterogeneous-pods/polaris-scheduler-75ms-5threads.jmx"
