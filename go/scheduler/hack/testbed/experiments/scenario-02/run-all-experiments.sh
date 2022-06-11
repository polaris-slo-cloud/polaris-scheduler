#!/bin/bash

# IMPORTANT: Configure the names of the scheduler pods here.
# Use the following syntax: "<namespace>:<pod-name"
# E.g., ["polaris-scheduler"]="polaris:polaris-scheduler-xyz"
declare -A schedulerPods=(
    ["greedy-first-fit-scheduler"]=""
    ["round-robin-scheduler"]=""
    ["timed-kube-scheduler"]=""
    ["polaris-scheduler"]=""
)

# The number of iterations to execute.
iterationsCount=5

# Specifies how many instances of the test app to deploy and what multiplier to use for the replica counts within an instance.
# E.g., "2x4" indicates 2 distinct instances of the test app (in 2 namespaces) and within each instance the replica multiplier 4 is used
# to multiply the base replica counts of the services.
# Base replica counts:
# collector: 3
# aggregator: 1
# hazard-broadcaster: 1
# traffic-info-provider: 1
# region-manager: 1 (not affected by multiplier)
instancesConfigs=(
    "1:1"
    "1:10"
    # "2:5"
)

experimentFiles=(
    "./test-app.template.yaml"
)

LOGS_DIR="./results/logs"

for scheduler in "${!schedulerPods[@]}"; do
    if [ "${schedulerPods[$scheduler]}" == "" ]; then
        echo "Please set the names of the scheduler pods in the schedulerPods array!"
        exit 1
    fi
done

mkdir -p "${LOGS_DIR}"

for scheduler in "${!schedulerPods[@]}"; do
    for instanceConfigStr in "${instancesConfigs[@]}"; do
        readarray -d ":" -t config <<< "${instanceConfigStr}"
        instanceCount=$(tr -d "\n" <<< ${config[0]})
        replicaMultiplier=$(tr -d "\n" <<< ${config[1]})

        for experimentFile in "${experimentFiles[@]}"; do
            startInfo="$(date) Running experiment $experimentFile with $scheduler and $instanceCount instances and replica multiplier $replicaMultipler"
            echo "$startInfo"

            ./run-single-experiment.sh "${experimentFile}" "${scheduler}" "${instanceCount}" "${replicaMultiplier}" "${iterationsCount}"

            readarray -d : -t podId <<< "${schedulerPods[$scheduler]}"
            podNamespace="${podId[0]}"
            podName=$(echo "${podId[1]}" | tr -d "\n")
            logFile="${LOGS_DIR}/${scheduler}-${instanceCount}instances-${replicaMultiplier}replicamult-$(date +%s).log"

            if [ "$podName" != "" ]; then
                echo "$startInfo" > "$logFile"
                kubectl logs -n "$podNamespace" "$podName" >> "$logFile"
            fi
        done
    done
done
