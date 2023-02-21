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
    # 2-smart sampling variations
    ["2-smart-10p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/10p-clusters-4p-samples.config.sh"
    ["2-smart-10p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/10p-clusters-8p-samples.config.sh"
    ["2-smart-10p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/10p-clusters-12p-samples.config.sh"
    ["2-smart-10p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/10p-clusters-16p-samples.config.sh"
    ["2-smart-20p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/20p-clusters-4p-samples.config.sh"
    ["2-smart-20p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/20p-clusters-8p-samples.config.sh"
    ["2-smart-20p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/20p-clusters-12p-samples.config.sh"
    ["2-smart-20p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/20p-clusters-16p-samples.config.sh"
    ["2-smart-30p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/30p-clusters-4p-samples.config.sh"
    ["2-smart-30p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/30p-clusters-8p-samples.config.sh"
    ["2-smart-30p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/30p-clusters-12p-samples.config.sh"
    ["2-smart-30p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/30p-clusters-16p-samples.config.sh"
    ["2-smart-40p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/40p-clusters-4p-samples.config.sh"
    ["2-smart-40p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/40p-clusters-8p-samples.config.sh"
    ["2-smart-40p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/40p-clusters-12p-samples.config.sh"
    ["2-smart-40p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/40p-clusters-16p-samples.config.sh"
    ["2-smart-50p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/50p-clusters-4p-samples.config.sh"
    ["2-smart-50p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/50p-clusters-8p-samples.config.sh"
    ["2-smart-50p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/50p-clusters-12p-samples.config.sh"
    ["2-smart-50p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/50p-clusters-16p-samples.config.sh"
    ["2-smart-60p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/60p-clusters-4p-samples.config.sh"
    ["2-smart-60p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/60p-clusters-8p-samples.config.sh"
    ["2-smart-60p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/60p-clusters-12p-samples.config.sh"
    ["2-smart-60p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/60p-clusters-16p-samples.config.sh"
    ["2-smart-70p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/70p-clusters-4p-samples.config.sh"
    ["2-smart-70p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/70p-clusters-8p-samples.config.sh"
    ["2-smart-70p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/70p-clusters-12p-samples.config.sh"
    ["2-smart-70p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/70p-clusters-16p-samples.config.sh"
    ["2-smart-80p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/80p-clusters-4p-samples.config.sh"
    ["2-smart-80p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/80p-clusters-8p-samples.config.sh"
    ["2-smart-80p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/80p-clusters-12p-samples.config.sh"
    ["2-smart-80p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/80p-clusters-16p-samples.config.sh"
    ["2-smart-90p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/90p-clusters-4p-samples.config.sh"
    ["2-smart-90p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/90p-clusters-8p-samples.config.sh"
    ["2-smart-90p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/90p-clusters-12p-samples.config.sh"
    ["2-smart-90p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/90p-clusters-16p-samples.config.sh"
    ["2-smart-100p-clusters-4p-samples"]="./experiments/2-smart-sampling-variations/100p-clusters-4p-samples.config.sh"
    ["2-smart-100p-clusters-8p-samples"]="./experiments/2-smart-sampling-variations/100p-clusters-8p-samples.config.sh"
    ["2-smart-100p-clusters-12p-samples"]="./experiments/2-smart-sampling-variations/100p-clusters-12p-samples.config.sh"
    ["2-smart-100p-clusters-16p-samples"]="./experiments/2-smart-sampling-variations/100p-clusters-16p-samples.config.sh"

    # 2-smart sampling variations LRU node
    ["2-smart-10p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/10p-clusters-4p-samples.config.sh"
    ["2-smart-10p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/10p-clusters-8p-samples.config.sh"
    ["2-smart-10p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/10p-clusters-12p-samples.config.sh"
    ["2-smart-10p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/10p-clusters-16p-samples.config.sh"
    ["2-smart-20p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/20p-clusters-4p-samples.config.sh"
    ["2-smart-20p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/20p-clusters-8p-samples.config.sh"
    ["2-smart-20p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/20p-clusters-12p-samples.config.sh"
    ["2-smart-20p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/20p-clusters-16p-samples.config.sh"
    ["2-smart-30p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/30p-clusters-4p-samples.config.sh"
    ["2-smart-30p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/30p-clusters-8p-samples.config.sh"
    ["2-smart-30p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/30p-clusters-12p-samples.config.sh"
    ["2-smart-30p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/30p-clusters-16p-samples.config.sh"
    ["2-smart-40p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/40p-clusters-4p-samples.config.sh"
    ["2-smart-40p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/40p-clusters-8p-samples.config.sh"
    ["2-smart-40p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/40p-clusters-12p-samples.config.sh"
    ["2-smart-40p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/40p-clusters-16p-samples.config.sh"
    ["2-smart-50p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/50p-clusters-4p-samples.config.sh"
    ["2-smart-50p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/50p-clusters-8p-samples.config.sh"
    ["2-smart-50p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/50p-clusters-12p-samples.config.sh"
    ["2-smart-50p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/50p-clusters-16p-samples.config.sh"
    ["2-smart-60p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/60p-clusters-4p-samples.config.sh"
    ["2-smart-60p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/60p-clusters-8p-samples.config.sh"
    ["2-smart-60p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/60p-clusters-12p-samples.config.sh"
    ["2-smart-60p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/60p-clusters-16p-samples.config.sh"
    ["2-smart-70p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/70p-clusters-4p-samples.config.sh"
    ["2-smart-70p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/70p-clusters-8p-samples.config.sh"
    ["2-smart-70p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/70p-clusters-12p-samples.config.sh"
    ["2-smart-70p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/70p-clusters-16p-samples.config.sh"
    ["2-smart-80p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/80p-clusters-4p-samples.config.sh"
    ["2-smart-80p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/80p-clusters-8p-samples.config.sh"
    ["2-smart-80p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/80p-clusters-12p-samples.config.sh"
    ["2-smart-80p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/80p-clusters-16p-samples.config.sh"
    ["2-smart-90p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/90p-clusters-4p-samples.config.sh"
    ["2-smart-90p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/90p-clusters-8p-samples.config.sh"
    ["2-smart-90p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/90p-clusters-12p-samples.config.sh"
    ["2-smart-90p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/90p-clusters-16p-samples.config.sh"
    ["2-smart-100p-clusters-4p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/100p-clusters-4p-samples.config.sh"
    ["2-smart-100p-clusters-8p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/100p-clusters-8p-samples.config.sh"
    ["2-smart-100p-clusters-12p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/100p-clusters-12p-samples.config.sh"
    ["2-smart-100p-clusters-16p-samples-lru-node"]="./experiments/2-smart-sampling-variations-lru-node/100p-clusters-16p-samples.config.sh"

    # Test scalability by varying jobs/sec on a 20K nodes cluster.
    ["scalability-jobs-01"]="./experiments/scalability-jobs/iteration-01.config.sh"
    ["scalability-jobs-02"]="./experiments/scalability-jobs/iteration-02.config.sh"
    ["scalability-jobs-03"]="./experiments/scalability-jobs/iteration-03.config.sh"
    ["scalability-jobs-04"]="./experiments/scalability-jobs/iteration-04.config.sh"
    ["scalability-jobs-05"]="./experiments/scalability-jobs/iteration-05.config.sh"
    ["scalability-jobs-06"]="./experiments/scalability-jobs/iteration-06.config.sh"
    ["scalability-jobs-07"]="./experiments/scalability-jobs/iteration-07.config.sh"
    ["scalability-jobs-08"]="./experiments/scalability-jobs/iteration-08.config.sh"
    ["scalability-jobs-09"]="./experiments/scalability-jobs/iteration-09.config.sh"
    ["scalability-jobs-10"]="./experiments/scalability-jobs/iteration-10.config.sh"
    ["scalability-jobs-11"]="./experiments/scalability-jobs/iteration-11.config.sh"
    ["scalability-jobs-12"]="./experiments/scalability-jobs/iteration-12.config.sh"
    ["scalability-jobs-13"]="./experiments/scalability-jobs/iteration-13.config.sh"
    ["scalability-jobs-14"]="./experiments/scalability-jobs/iteration-14.config.sh"
    ["scalability-jobs-15"]="./experiments/scalability-jobs/iteration-15.config.sh"

    # Test scalability by varying cluster size for scheduling 1000 pods.
    # We run each config 3 times.
    ["scalability-nodes-01"]="./experiments/scalability-nodes/1k.config.sh"
    ["scalability-nodes-02"]="./experiments/scalability-nodes/1k.config.sh"
    ["scalability-nodes-03"]="./experiments/scalability-nodes/1k.config.sh"
    ["scalability-nodes-04"]="./experiments/scalability-nodes/5k.config.sh"
    ["scalability-nodes-05"]="./experiments/scalability-nodes/5k.config.sh"
    ["scalability-nodes-06"]="./experiments/scalability-nodes/5k.config.sh"
    ["scalability-nodes-07"]="./experiments/scalability-nodes/10k.config.sh"
    ["scalability-nodes-08"]="./experiments/scalability-nodes/10k.config.sh"
    ["scalability-nodes-09"]="./experiments/scalability-nodes/10k.config.sh"
    ["scalability-nodes-10"]="./experiments/scalability-nodes/15k.config.sh"
    ["scalability-nodes-11"]="./experiments/scalability-nodes/15k.config.sh"
    ["scalability-nodes-12"]="./experiments/scalability-nodes/15k.config.sh"
    ["scalability-nodes-13"]="./experiments/scalability-nodes/20k.config.sh"
    ["scalability-nodes-14"]="./experiments/scalability-nodes/20k.config.sh"
    ["scalability-nodes-15"]="./experiments/scalability-nodes/20k.config.sh"

    # Test scalability with pure-random sampling (i.e., no 2-smart sampling) by varying jobs/sec on a 20K nodes cluster.
    ["scalability-jobs-pure-random-01"]="./experiments/scalability-jobs-pure-random/iteration-01.config.sh"
    ["scalability-jobs-pure-random-02"]="./experiments/scalability-jobs-pure-random/iteration-02.config.sh"
    ["scalability-jobs-pure-random-03"]="./experiments/scalability-jobs-pure-random/iteration-03.config.sh"
    ["scalability-jobs-pure-random-04"]="./experiments/scalability-jobs-pure-random/iteration-04.config.sh"
    ["scalability-jobs-pure-random-05"]="./experiments/scalability-jobs-pure-random/iteration-05.config.sh"
    ["scalability-jobs-pure-random-06"]="./experiments/scalability-jobs-pure-random/iteration-06.config.sh"
    ["scalability-jobs-pure-random-07"]="./experiments/scalability-jobs-pure-random/iteration-07.config.sh"
    ["scalability-jobs-pure-random-08"]="./experiments/scalability-jobs-pure-random/iteration-08.config.sh"
    ["scalability-jobs-pure-random-09"]="./experiments/scalability-jobs-pure-random/iteration-09.config.sh"
    ["scalability-jobs-pure-random-10"]="./experiments/scalability-jobs-pure-random/iteration-10.config.sh"
    ["scalability-jobs-pure-random-11"]="./experiments/scalability-jobs-pure-random/iteration-11.config.sh"
    ["scalability-jobs-pure-random-12"]="./experiments/scalability-jobs-pure-random/iteration-12.config.sh"
    ["scalability-jobs-pure-random-13"]="./experiments/scalability-jobs-pure-random/iteration-13.config.sh"
    ["scalability-jobs-pure-random-14"]="./experiments/scalability-jobs-pure-random/iteration-14.config.sh"
    ["scalability-jobs-pure-random-15"]="./experiments/scalability-jobs-pure-random/iteration-15.config.sh"
)


# Pre-computed absolute paths on the various VMs. Do not modify.
TESTBED_PATH_NODE_VMS="$POLARIS_SCHED_REPO_NODE_VMS/$TESTBED_PATH_IN_REPO"
TESTBED_PATH_SCHEDULER_VM="$POLARIS_SCHED_REPO_SCHEDULER_VM/$TESTBED_PATH_IN_REPO"
SCRIPTS_ROOT_NODE_VMS="$TESTBED_PATH_NODE_VMS/$SCRIPTS_ROOT"
SCRIPTS_ROOT_SCHEDULER_VM="$TESTBED_PATH_SCHEDULER_VM/$SCRIPTS_ROOT"
RESULTS_ROOT_NODE_VMS="$TESTBED_PATH_NODE_VMS/$RESULTS_ROOT"
RESULTS_ROOT_SCHEDULER_VM="$TESTBED_PATH_SCHEDULER_VM/$RESULTS_ROOT"
