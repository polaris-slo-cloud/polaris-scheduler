#!/bin/bash

# To change the Kubernetes version, either set the MICRO_K8S_CHANNEL environment variable or modify the assignment below.
if [ "$MICRO_K8S_CHANNEL" == "" ]; then
    MICRO_K8S_CHANNEL="1.25/stable"
fi

# The path of the jmeter.sh script (only used on the load generator VM).
JMETER_SH="$HOME/apache-jmeter-5.5/bin/jmeter.sh"

# The path of the polaris-scheduler git repository on the cluster node VMs.
# Note that the '$' of env vars (e.g., $HOME) must be escaped, because this path
# will included in an ssh command and the env vars must resolve on the remote VM, not on the local machine.
POLARIS_SCHED_REPO_NODE_VMS="\$HOME/polaris-scheduler"

# The path of the polaris-scheduler git repository on the VM that hosts the scheduler.
# Note that the '$' of env vars (e.g., $HOME) must be escaped, because this path
# will included in an ssh command and the env vars must resolve on the remote VM, not on the local machine.
POLARIS_SCHED_REPO_SCHEDULER_VM="\$HOME/polaris-scheduler"

# The path of the testbed folder, relative to the polaris-scheduler git repo.
TESTBED_PATH_IN_REPO="./testbed/automated-multi-vm-fake-kubelet-multi-cluster"

# The directory, relative to the testbed folder, where to store the resulting log files.
RESULTS_ROOT="./results"

# The scripts directory, relative to the testbed folder.
SCRIPTS_ROOT="./scripts"

# The list of SSH destination strings used to connect to the VMs that host the experiment's clusters.
CLUSTER_VMS=(
    "user@1.2.3.4"
)

# The SSH destination string to connect to the VM that hosts the polaris-scheduler.
SCHEDULER_VM="user@1.2.3.4"

# The local port that should be forwarded to the scheduler VM to allow connecting to the scheduler's REST API.
# This must be the same port that is used in the JMeter test plans.
SCHEDULER_LOCAL_PORT="38080"

# The port on the SCHEDULER_VM to which the SCHEDULER_LOCAL_PORT should be forwarded.
# This must be the same as the por configured in the polaris-scheduler docker-compose.yml files.
SCHEDULER_REMOTE_PORT="38080"

# (optional) Configure non-standard SSH ports for CLUSTER_VMS and SCHEDULER_VM.
# Use the SSH destination string as a key in this dictionary to configure a non-standard port.
declare -A SSH_PORTS=(
    # ["user@1.2.3.4"]="22"
)

# The list of experiment iteration config files, relative to the run-experiments.sh file.
declare -A EXPERIMENT_ITERATIONS=(
    ["scalability-jobs-per-sec-01"]="./experiments/scalability-jobs-per-sec/iteration-01.config.sh"
    ["scalability-jobs-per-sec-02"]="./experiments/scalability-jobs-per-sec/iteration-02.config.sh"
)


# Pre-computed absolute paths on the various VMs. Do not modify.
TESTBED_PATH_NODE_VMS="$POLARIS_SCHED_REPO_NODE_VMS/$TESTBED_PATH_IN_REPO"
TESTBED_PATH_SCHEDULER_VM="$POLARIS_SCHED_REPO_SCHEDULER_VM/$TESTBED_PATH_IN_REPO"
SCRIPTS_ROOT_NODE_VMS="$TESTBED_PATH_NODE_VMS/$SCRIPTS_ROOT"
SCRIPTS_ROOT_SCHEDULER_VM="$TESTBED_PATH_SCHEDULER_VM/$SCRIPTS_ROOT"
RESULTS_ROOT_NODE_VMS="$TESTBED_PATH_NODE_VMS/$RESULTS_ROOT"
RESULTS_ROOT_SCHEDULER_VM="$TESTBED_PATH_SCHEDULER_VM/$RESULTS_ROOT"
