#!/bin/bash
# set -x
set -o errexit
set -m

# This script runs Apache JMeter to generate load on the scheduler.
# Parameters:
# $1 - the absolute path to an Apache JMeter test plan file (*.jmx).
# $2 - the absolute path to the results directory to be used for this experiment iteration.

###############################################################################
# Global variables and imports
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

JMETER_TEST_PLAN="$1"
RESULTS_DIR="$2"
ITERATION_NAME="$3"


###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "This script runs Apache JMeter to generate load on the scheduler."
    echo "Usage:"
    echo "./scripts/04-run-load-generator.sh <jmeter-test-plan.jmx (absolute path)> <result directory (absolute path)> <iteration name>"
    echo "Example: "
    echo "./scripts/04-run-load-generator.sh \"\$(pwd)/jmeter-test-plans/heterogeneous-pods/polaris-scheduler-10ms-5threads.jmx\" \"\$(pwd)/results\" \"iteration-01\""
}

function validateJMeterTestPlanPath() {
    if [ "$JMETER_TEST_PLAN" == "" ] || [ ! -f "$JMETER_TEST_PLAN" ]; then
        printError "Please specify the absolute path of an Apache JMeter test plan file (*.jmx) as the first argument."
        printUsage
        exit 1
    fi
}

function validateResultsDir() {
    if [ "$RESULTS_DIR" == "" ] || [ ! -d "$RESULTS_DIR" ]; then
        printError "Please specify the absolute path of an existing directory for storing the results as the second argument."
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

validateJMeterTestPlanPath
validateResultsDir
validateIterationName

echo "Running JMeter with test plan $JMETER_TEST_PLAN"
logFile="$RESULTS_DIR/jmeter-$ITERATION_NAME.log"

("$JMETER_SH" -n -t "$JMETER_TEST_PLAN" -j "$logFile")

echo "Successfully completed JMeter test."

