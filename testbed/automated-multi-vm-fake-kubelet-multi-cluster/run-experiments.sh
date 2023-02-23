#!/bin/bash
# set -x
set -o errexit
set -m

# This script runs the experiments configured in experiment.config.sh in a fully automated fashion.
# The requirements are:
# - 11 VMs (10 for running the simulated clusters, 1 for the scheduler)
# - SSH access to all VMs configured in the config file.
# - passwordless sudo on the CLUSTER VMs.
#
# IMPORTANT: Before starting the script, ensure that the polaris-scheduler-config.yaml files contain the correct
# addresses of the remoteClusters.

###############################################################################
# Global variables
# These must not start with an underscore to avoid clashing with
# locally used variables that need to be declared globally (using declare).
###############################################################################

# DEBUG=1

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))

START_SINGLE_CLUSTER_SCRIPT="${SCRIPT_DIR}/../fake-kubelet-cluster/start-kind-with-fake-kubelet.sh"
CLUSTER_AGENT_DEPLOYMENT_YAML="${SCRIPT_DIR}/polaris-cluster-agent"

CLUSTER_AGENT_NODE_PORT=30033


###############################################################################
# Functions
###############################################################################

# Checks the experiment iteration files.
function checkExperimentIterationFiles() {
    logMsg "Checking experiment configuration files"
    local _iterationFile=""
    for _iterationName in "${!EXPERIMENT_ITERATIONS[@]}"; do
        (
            _iterationFile="$SCRIPT_DIR/${EXPERIMENT_ITERATIONS[$_iterationName]}"
            if [ "$_iterationFile" == "$SCRIPT_DIR" ] || [ ! -f "$_iterationFile" ]; then
                printError "$_iterationName: The experiment iteration file \"$_iterationFile\" could not be found."
                exit 1
            fi
            source "$_iterationFile"

            if [ ! -d "$SCRIPT_DIR/$SCHEDULER_DOCKER_COMPOSE_DIR" ]; then
                printError "$_iterationName: The SCHEDULER_DOCKER_COMPOSE_DIR could not be found"
                exit 1
            fi
            if [ ! -d "$SCRIPT_DIR/$CLUSTER_AGENT_DEPLOYMENT_YAML_DIR" ]; then
                printError "$_iterationName: The CLUSTER_AGENT_DEPLOYMENT_YAML_DIR could not be found"
                exit 1
            fi
            if [ ! -f "$SCRIPT_DIR/$JMETER_TEST_PLAN" ]; then
                printError "$_iterationName: The JMETER_TEST_PLAN file could not be found"
                exit 1
            fi

            for clusterConfig in "${CLUSTER_CONFIGS[@]}"; do
                if [ ! -f "$SCRIPT_DIR/$clusterConfig" ]; then
                    printError "$_iterationName: The cluster config $clusterConfig could not be found"
                    exit 1
                fi
            done
        )
    done
}

# Checks the SSH connections to all VMs
function checkSshConnections() {
    logMsg "Checking remote VMs for SSH access"
    sshRunCmd "$SCHEDULER_VM" "hostname"
    sshRunCmdOnMultipleSystems "CLUSTER_VMS" "hostname"
}

# Forwards the local port to the scheduler VM.
# Sets $RET to the PID of the ssh process.
function forwardPortToSchedulerVM() {
    logMsg "Forwarding local port $SCHEDULER_LOCAL_PORT to port $SCHEDULER_REMOTE_PORT on the scheduler VM to allow access to the scheduler's REST API."
    ssh -N -L $SCHEDULER_LOCAL_PORT:localhost:$SCHEDULER_REMOTE_PORT $SCHEDULER_VM &
    RET=$!
}

# Starts the fake-kubelet clusters on the cluster VMs.
function startClusters() {
    local clusterVm=""
    local configFile=""
    local length="${#CLUSTER_VMS[@]}"

    for (( i=0; i<${length}; i++ )); do
        clusterVm="${CLUSTER_VMS[$i]}"
        configFile="${CLUSTER_CONFIGS[$i]}"

        sshRunCmd "$clusterVm" "\"$SCRIPTS_ROOT_NODE_VMS/02-start-cluster.sh\" \"$TESTBED_PATH_NODE_VMS/$configFile\" \"$TESTBED_PATH_NODE_VMS/$CLUSTER_AGENT_DEPLOYMENT_YAML_DIR\"" &
    done

    # Wait for all commands to complete.
    wait $(jobs -p)
}

# Runs a single experiment iteration. This MUST be run in a subshell.
# Parameters:
# $1 - the name of the experiment iteration
# $2 - the experiment counter (used as an informative counter only)
function runExperimentIteration() {
    local iterationName="$1"
    local experimentCounter="$2"
    local iterationFile=${EXPERIMENT_ITERATIONS[$iterationName]}
    iterationFile="$SCRIPT_DIR/$iterationFile"
    logMsg "Running experiment iteration: $iterationName ($experimentCounter of ${#EXPERIMENT_ITERATIONS[@]})"
    if [ "$iterationFile" == "$SCRIPT_DIR" ] || [ ! -f "$iterationFile" ]; then
        echo "The experiment iteration file \"$iterationFile\" could not be found."
        exit 1
    fi

    local iterationStart="$(date +%s)"
    source "$iterationFile"

    logMsg "$iterationName Step 1: Installing MicroK8s on ${#CLUSTER_VMS[@]} cluster VMs"
    sshRunCmdOnMultipleSystems "CLUSTER_VMS" "sudo bash -c \"$SCRIPTS_ROOT_NODE_VMS/01-install-microk8s.sh\""

    logMsg "$iterationName Step 2: Starting fake-kubelet clusters"
    startClusters

    logMsg "$iterationName Step 3: Starting polaris-scheduler"
    sshRunCmd "$SCHEDULER_VM" "\"$SCRIPTS_ROOT_SCHEDULER_VM/03-start-polaris-scheduler.sh\" \"$TESTBED_PATH_SCHEDULER_VM/$SCHEDULER_DOCKER_COMPOSE_DIR\""

    logMsg "$iterationName Step 4: Run load generator"
    ("$SCRIPT_DIR/$SCRIPTS_ROOT/04-run-load-generator.sh" "$SCRIPT_DIR/$JMETER_TEST_PLAN" "$SCRIPT_DIR/$RESULTS_ROOT" "$iterationName")

    logMsg "$iterationName Step 5: Waiting for experiment to finish"
    sshRunCmd "$SCHEDULER_VM" "\"$SCRIPTS_ROOT_SCHEDULER_VM/05-wait-until-experiment-done.sh\" \"$TESTBED_PATH_SCHEDULER_VM/$SCHEDULER_DOCKER_COMPOSE_DIR\" $LONG_SLEEP"

    logMsg "$iterationName Step 6: Exporting scheduler logs"
    sshRunCmd "$SCHEDULER_VM" "\"$SCRIPTS_ROOT_SCHEDULER_VM/06-export-scheduler-logs-and-stop-scheduler.sh\" \"$TESTBED_PATH_SCHEDULER_VM/$SCHEDULER_DOCKER_COMPOSE_DIR\" \"$RESULTS_ROOT_SCHEDULER_VM\" \"$iterationName\""

    logMsg "$iterationName Step 7: Exporting cluster agent logs"
    sshRunCmdOnMultipleSystems "CLUSTER_VMS" "\"$SCRIPTS_ROOT_NODE_VMS/07-export-cluster-agent-logs.sh\" \"$RESULTS_ROOT_NODE_VMS\" \"$iterationName\""

    logMsg "$iterationName Step 8: Uninstalling MicroK8s from ${#CLUSTER_VMS[@]} cluster VMs"
    sshRunCmdOnMultipleSystems "CLUSTER_VMS" "sudo bash -c \"$SCRIPTS_ROOT_NODE_VMS/08-uninstall-microk8s.sh\""

    local iterationStop="$(date +%s)"
    local duration=$(($iterationStop - $iterationStart))
    logMsg "$iterationName Finished experiment in $duration seconds"
    logMsg ""
}

# Copies the results from all remote VMs.
function copyResults() {
    logMsg "Copying results"

    scp -r "$SCHEDULER_VM:$RESULTS_ROOT_SCHEDULER_VM" "$SCRIPT_DIR/$RESULTS_ROOT/scheduler"

    local length="${#CLUSTER_VMS[@]}"
    local clusterVm=""
    local clusterName=""
    for (( i=0; i<${length}; i++ )); do
        clusterVm="${CLUSTER_VMS[$i]}"

        clusterName=$((i+1))
        clusterName=$(printf "cluster-%02d" $clusterName)
        scp -r "$clusterVm:$RESULTS_ROOT_NODE_VMS" "$SCRIPT_DIR/$RESULTS_ROOT/$clusterName"
    done
}


###############################################################################
# Script Start
###############################################################################

source "$SCRIPT_DIR/experiment.config.sh"
source "$SCRIPT_DIR/scripts/lib/util.sh"

checkExperimentIterationFiles
checkSshConnections

forwardPortToSchedulerVM
portForwardingSshPid=$RET

(
    # Run the experiments in a subshell to ensure that waiting for all child processes does not include the port forwarding SSH process.

    logMsg "Starting experiments"
    logMsg ""
    expStartTime="$(date +%s)"
    expCounter="0"

    mkdir -p "$SCRIPT_DIR/$RESULTS_ROOT"

    for expIterationName in "${!EXPERIMENT_ITERATIONS[@]}"; do
        expCounter=$(( $expCounter+1))
        (runExperimentIteration "$expIterationName" "$expCounter")
    done

    copyResults

    expEndTime="$(date +%s)"
    duration=$(($expEndTime - $expStartTime))
    logMsg "Ran ${#EXPERIMENT_ITERATIONS[@]} experiments in $duration seconds."
)

logMsg "Stopping port forwarding to scheduler VM"
kill -SIGTERM $portForwardingSshPid
