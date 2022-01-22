#!/bin/bash

# IMPORTANT: Configure the names of the scheduler pods here.
# Use the following syntax: "<namespace>:<pod-name"
# E.g., ["rainbow-scheduler"]="rainbow-system:rainbow-scheduler-xyz"
declare -A schedulerPods=(
    ["greedy-first-fit-scheduler"]=""
    ["round-robin-scheduler"]=""
    ["timed-kube-scheduler"]=""
    ["rainbow-scheduler"]=""
)

# The number of iterations to execute.
iterationsCount=5

instanceCounts=(
    "1"
    # "8"
)

experimentFiles=(
    "./test-app.template.yaml"
)

for scheduler in "${!schedulerPods[@]}"; do
    if [ "${schedulerPods[$scheduler]}" == "" ]; then
        echo "Please set the names of the scheduler pods in the schedulerPods array!"
        exit 1
    fi
done

mkdir -p "logs"

for scheduler in "${!schedulerPods[@]}"; do
    for instanceCount in "${instanceCounts[@]}"; do
        for experimentFile in "${experimentFiles[@]}"; do
            startInfo="$(date) Running experiment $experimentFile with $scheduler and $instanceCount instances"
            echo "$startInfo"

            ./run-single-experiment.sh "${experimentFile}" "${scheduler}" "${instanceCount}" "${iterationsCount}"

            readarray -d : -t podId <<< "${schedulerPods[$scheduler]}"
            podNamespace="${podId[0]}"
            podName=$(echo "${podId[1]}" | tr -d "\n")
            logFile="./logs/${scheduler}-${instanceCount}instances-$(date +%s).log"

            if [ "$podName" != "" ]; then
                echo "$startInfo" > "$logFile"
                kubectl logs -n "$podNamespace" "$podName" >> "$logFile"
            fi
        done
    done
done
