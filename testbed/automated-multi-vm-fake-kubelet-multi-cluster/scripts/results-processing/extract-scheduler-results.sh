#!/bin/bash
# set -x
set -o errexit
set -m

# This script extracts various results from the scheduler log files.
# Parameters:
# $1 - Path of the folder that contains the scheduler .log files.

###############################################################################
# Global variables
# These must not start with an underscore to avoid clashing with
# locally used variables that need to be declared globally (using declare).
###############################################################################

# DEBUG=1

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))

RESULTS_DIR="$1"
OUT_FILE="$2"

CURR_LINE=""

###############################################################################
# Functions
###############################################################################

function printUsage() {
    echo "This script extracts various results from the JMeter and scheduler log files."
    echo "Usage:"
    echo "./scripts/extract-scheduler-results.sh <results directory> <csv output directory>"
    echo "Example: "
    echo "./scripts/extract-scheduler-results.sh ./scheduler-results ./output.csv"
}

function validateResultsDir() {
    if [ "$RESULTS_DIR" == "" ] || [ ! -d "$RESULTS_DIR" ]; then
        printError "Please specify the path of the directory that contains the results files from the experiments as the first argument."
        printUsage
        exit 1
    fi
}

function ensureJqInstalled() {
    if [ "$(which jq)" == "" ]; then
        printError "Please ensure that 'jq' (Command-line JSON processor) is installed. This is required for computing average statistics."
        exit 1
    fi
}

function validateOutFile() {
    if [ "$OUT_FILE" == "" ]; then
        printError "Please specify the output CSV file as the second argument."
        printUsage
        exit 1
    fi
}

function writeHeaderRow() {
    local headerRow='"Experiment","Total Pods","Scheduling Successes","Scheduling Failures (incl. retries)","Scheduling Failures (final - no more retries)","Scheduling Conflicts","Scheduling Conflicts if no MultiBinding","Avg queuing time (successes and failures)","Avg sampling duration (successes and failures)","Avg sampled nodes","Avg eligible nodes","Avg commit duration (successes)","Avg E2E duration (successes)","First Successful Pod Timestamp","Last Successful Pod Timestamp"'
    echo "$headerRow" > "$OUT_FILE"
}

# Extracts the experiment name from the log file path in $1 and stores it in $RET
function extractExperimentName() {
    local fileName=$(basename "$1")
    local expName="${fileName#*-}"
    expName="${expName%.*}"
    RET="$expName"
}

# Gets the total number of submitted pods from the JMeter log file specified in $1 and stores the result in $RET.
function getTotalPods() {
    local jmeterLogFile=$1
    RET=$(cat "$jmeterLogFile" | awk '{match($0, /^.+summary =\s+([0-9]+)\s.+/, arr); print arr[1];}' | tail -n 1)
}

# Counts the matches of the regex in $1 within the file $2 and stores the result in $RET
function countMatches() {
    local regexStr=$1
    local filePath=$2
    RET=$(grep -E -c "$regexStr" "$filePath" || true)
}

# Counts the matches of the regex in $1 within the file $2, stores the result it $RET and appends it
# as a CSV column to $CURR_LINE
function countAndAppendMatches() {
    local regexStr=$1
    local filePath=$2
    countMatches "$regexStr" "$filePath"
    CURR_LINE="${CURR_LINE},\"$RET\""
}

# Extracts the average sampling duration from the scheduler log file in $1 and stores the result in $RET.
function extractAvgSamplingDuration() {
    local schedulerLogFile=$1
    # Note about jq: -s (--slurp) creates an array for the input lines after parsing each line as JSON, or as a number in this case.
    RET=$(cat "$schedulerLogFile" | awk '{match($0, /^.+samplingDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | jq -s add/length)
}

# Extracts the average number of sampled nodes from the scheduler log file in $1 and stores the result in $RET.
function extractAvgSampledNodes() {
    local schedulerLogFile=$1
    RET=$(cat "$schedulerLogFile" | awk '{match($0, /^.+sampledNodes"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | jq -s add/length)
}

# Extracts the average number of eligible nodes from the scheduler log file in $1 and stores the result in $RET.
function extractAvgEligibleNodes() {
    local schedulerLogFile=$1
    RET=$(cat "$schedulerLogFile" | awk '{match($0, /^.+eligibleNodes"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | jq -s add/length)
}

# Extracts the queuing time from the scheduler log file in $1 and stores the result in $RET.
function extractAvgQueuingTime() {
    local schedulerLogFile=$1
    RET=$(cat "$schedulerLogFile" | awk '{match($0, /^.+queueTimeMs"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | jq -s add/length)
}

# Extracts the commit duration (successes only) from the scheduler log file in $1 and stores the result in $RET.
function extractAvgCommitDuration() {
    local schedulerLogFile=$1
    RET=$(grep -E '"SchedulingSuccess"' "$schedulerLogFile" | awk '{match($0, /^.+commitDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' | jq -s add/length)
}

# Extracts the E2E duration (successes only) from the scheduler log file in $1 and stores the result in $RET.
function extractAvgE2EDuration() {
    local schedulerLogFile=$1
    RET=$(grep -E '"SchedulingSuccess"' "$schedulerLogFile" | awk '{match($0, /^.+e2eDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' | jq -s add/length)
}

# Extracts the timestamp of the first pod from the scheduler log file in $1 and stores the result in $RET.
function extractFirstPodTimestamp() {
    local schedulerLogFile=$1
    RET=$(head -n 1000 "$schedulerLogFile" | awk '{match($0, /^.+unixTimestampMs"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | head -n 1)
}

# Extracts the timestamp of the last pod from the scheduler log file in $1 and stores the result in $RET.
function extractLastPodTimestamp() {
    local schedulerLogFile=$1
    RET=$(cat "$schedulerLogFile" | awk '{match($0, /^.+unixTimestampMs"=([0-9]+)\s.+/, arr); print arr[1];}' | grep -E '[0-9]+' | tail -n 1)
}

###############################################################################
# Script Start
###############################################################################

source "$SCRIPT_DIR/../lib/util.sh"

validateResultsDir
validateOutFile
ensureJqInstalled

allLogsUnsorted=($RESULTS_DIR/jmeter-*.log)
IFS=$'\n' allLogs=($(sort <<<"${allLogsUnsorted[*]}")); unset IFS

writeHeaderRow "$OUT_FILE"

processedCount=0
startTime=$(date +%s)

for jmeterLog in "${allLogs[@]}"; do
    extractExperimentName "$jmeterLog"
    expName=$RET
    echo "Processing $expName"

    schedulerLog="$(dirname "$jmeterLog")/scheduler/scheduler-${expName}.log"
    if [ ! -f "$schedulerLog" ]; then
        printError "Scheduler log file $schedulerLog does not exist. Skipping"
        continue
    fi

    CURR_LINE="\"$expName\""

    # Total Pods
    getTotalPods "$jmeterLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Scheduling Successes
    countAndAppendMatches '"SchedulingSuccess"' "$schedulerLog"

    # Scheduling Failures (incl. retries)
    countAndAppendMatches '"FailedScheduling".+"reason"=' "$schedulerLog"

    # Scheduling Failures (final - no more retries)
    countAndAppendMatches '"FailedScheduling".+"reason"=.+"retryingScheduling"=false' "$schedulerLog"

    # Scheduling Conflicts
    countAndAppendMatches '"FailedScheduling".+"reasons"=' "$schedulerLog"

    # Scheduling Conflicts if no MultiBinding
    countAndAppendMatches '"SchedulingSuccess".+"commitRetries"=[1-3]' "$schedulerLog"

    # Avg queuing time (successes and failures)
    extractAvgQueuingTime "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Avg sampling duration (successes and failures)
    extractAvgSamplingDuration "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Avg sampled nodes
    extractAvgSampledNodes "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Avg eligible nodes
    extractAvgEligibleNodes "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Avg commit duration (successes)
    extractAvgCommitDuration "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Avg E2E duration (successes)
    extractAvgE2EDuration "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # First Successful Pod Timestamp
    extractFirstPodTimestamp "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    # Last Successful Pod Timestamp
    extractLastPodTimestamp "$schedulerLog"
    CURR_LINE="${CURR_LINE},\"$RET\""

    echo "$CURR_LINE" >> "$OUT_FILE"

    processedCount=$((processedCount+1))
done

endTime=$(date +%s)
echo "Successfully processed $processedCount of ${#allLogs[@]} experiments in $((endTime - startTime)) seconds."
