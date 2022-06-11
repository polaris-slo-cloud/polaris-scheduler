#!/bin/bash
# set -x
# set -o errexit

function printUsage() {
    echo "Usage:"
    echo "./run-single-experiment.sh [deployment-YAML-template] [schedulerName] [instanceCount] [replicaMultiplier] [iterationsCount]"
    echo "Example with 2 instances, a replica multiplier of 4, 5 iterations, and polaris-scheduler:"
    echo "./run-single-experiment.sh ./test-app.template.yaml polaris-scheduler 2 4 5"
}

if [ "$1" == "" ] || [ ! -f $1 ]; then
    echo "Please provide the name of a deployment template YAML as the first argument."
    printUsage
    exit 1
fi

if [ "$2" == "" ]; then
    echo "Please provide the scheduler name as the second argument."
    printUsage
    exit 1
fi

if [ "$3" == "" ]; then
    echo "Please provide the instance count as the third argument."
    printUsage
    exit 1
fi

if [ "$4" == "" ]; then
    echo "Please provide the replica multiplier as the fourth argument."
    printUsage
    exit 1
fi

if [ "$5" == "" ]; then
    echo "Please provide the number of iterations as the fifth argument."
    printUsage
    exit 1
fi

deploymentConfigs=("$1")
schedulerName="$2"
instanceCount="$3"
replicaMultiplier="$4"
iterationsCount="$5"
namespaceBase="traffic-"
shortSleepTime="20s"
longSleepTime="1m"
resultsDirSuffix=""
resultsDir="results"
totalPods=3
deployedNamespaces=()
separatorLines="\n---------------------------------------------------------------------------\n"

let collectorReplicas=replicaMultiplier*3
let aggregatorReplicas=replicaMultiplier*1
let hazardBroadcasterReplicas=replicaMultiplier*1
let trafficInfoProviderReplicas=replicaMultiplier*1
let regionManagerReplicas=1

# Waits until all pods are ready and appends the checking output to the log file.
function waitForResult() {
    local deployment=$1
    local iteration=$2
    local outFile="${resultsDir}/iteration${iteration}.json"

    echo "$(date): kubectl apply complete."
    sleepVerbosely $longSleepTime
    echo -e "{\n  \"schedulerName\": \"${schedulerName}\",\n  \"deployment\": \"${deployment}\",\n  \"iteration\": ${iteration},\n  \"instanceCount\": ${instanceCount},\n  \"results\": [\n" > "$outFile"

    for n in "${!deployedNamespaces[@]}"; do
        local namespace=${deployedNamespaces[$n]}
        if [ $n -gt 0 ]; then
            echo -e ",\n" >> "$outFile"
        fi
        echo -e "{\n  \"namespace\": \"${namespace}\",\n    \"result\":\n" >> "$outFile"
        local output=$(kubectl get pods -n $namespace -o json)
        echo "$output" >> "$outFile"
        echo "Checking namespace: ${namespace}"
        echo -e "\n}" >> "$outFile"
    done

    echo -e "\n],  \"nodes\": " >> "$outFile"
    kubectl get nodes -o json >> "$outFile"
    echo -e "\n}" >> "$outFile"
}

function undeploy() {
    local namespace=$1
    echo "Deleting namespace $namespace"
    kubectl delete namespace $namespace

    local output=$(kubectl get pods -n $namespace -o wide)
    while [ -n "$output" ]; do
        sleepVerbosely $shortSleepTime
        output=$(kubectl get pods -n $namespace -o wide)
    done
    echo "Deletion of namespace ${namespace} verified"
}

function sleepVerbosely() {
    echo "Sleeping for $1"
    sleep $1
}

function executeIteration() {
    local deploymentTemplateFile=$1
    local iteration=$2
    deployedNamespaces=()
    echo "Deploying iteration ${deploymentTemplateFile}-${iteration}"

    local fullDeploymentYaml=""

    for instance in $(seq 1 $instanceCount); do
        local finalNamespace="${namespaceBase}${instance}"
        local deploymentYaml=$(sed \
            -e "s/{{ \.Namespace }}/${finalNamespace}/" \
            -e "s/{{ \.SchedulerName }}/${schedulerName}/" \
            -e "s/{{ \.CollectorReplicas }}/${collectorReplicas}/" \
            -e "s/{{ \.AggregatorReplicas }}/${aggregatorReplicas}/" \
            -e "s/{{ \.HazardBroadcasterReplicas }}/${hazardBroadcasterReplicas}/" \
            -e "s/{{ \.TrafficInfoProviderReplicas }}/${trafficInfoProviderReplicas}/" \
            -e "s/{{ \.RegionManagerReplicas }}/${regionManagerReplicas}/" \
            ${deploymentTemplateFile})
        fullDeploymentYaml=$(echo -e "${fullDeploymentYaml}\n---\n${deploymentYaml}")
        deployedNamespaces+=("${finalNamespace}")
    done

    kubectl apply -f - <<< "${fullDeploymentYaml}"
    waitForResult $deployment $iteration

    echo "Undeploying iteration ${iteration}"
    for namespace in "${deployedNamespaces[@]}"; do
        undeploy "$namespace"
    done
}


for deployment in "${deploymentConfigs[@]}"; do
    resultsDir="$(dirname "${deployment}")/${schedulerName}-${resultsDirSuffix}/$(basename ${deployment})-${instanceCount}instances"
    mkdir -p "$resultsDir"
    for i in $(seq 1 $iterationsCount ); do
        echo "$deployment iteration: $i"
        executeIteration "$deployment" $i
        echo -e "\n"
    done
done
