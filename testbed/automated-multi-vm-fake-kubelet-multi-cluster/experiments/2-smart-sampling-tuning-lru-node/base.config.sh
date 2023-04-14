#!/bin/bash

# IMPORTANT: All paths MUST be relative to $TESTBED_PATH_IN_REPO configured in the root experiment.config.sh file.

# The path of the directory containing the cluster agent deployment YAML files.
CLUSTER_AGENT_DEPLOYMENT_YAML_DIR="./polaris-cluster-agent/2-smart-sampling-lru-node"

# Paths of the 10 cluster configuration files.
CLUSTER_CONFIGS=(
    "./clusters/20k-nodes/cluster-01.config.sh"
    "./clusters/20k-nodes/cluster-02.config.sh"
    "./clusters/20k-nodes/cluster-03.config.sh"
    "./clusters/20k-nodes/cluster-04.config.sh"
    "./clusters/20k-nodes/cluster-05.config.sh"
    "./clusters/20k-nodes/cluster-06.config.sh"
    "./clusters/20k-nodes/cluster-07.config.sh"
    "./clusters/20k-nodes/cluster-08.config.sh"
    "./clusters/20k-nodes/cluster-09.config.sh"
    "./clusters/20k-nodes/cluster-10.config.sh"
)

# Path of the JMeter test plan file.
JMETER_TEST_PLAN="./jmeter-test-plans/4cpu-4gi-pods/polaris-scheduler-10ms-25threads-11200pods.jmx"
