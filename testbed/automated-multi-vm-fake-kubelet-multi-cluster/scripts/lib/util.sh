#!/bin/bash

###############################################################################
# Global Variables
# These must not start with an underscore to avoid clashing with
# locally used variables that need to be declared globally (using declare).
###############################################################################

RET=""

SHORT_SLEEP="10s"
MEDIUM_SLEEP="20s"
LONG_SLEEP="1m"

if [ "$DEBUG" == "" ]; then
    DEBUG=0
fi


###############################################################################
# Utility Functions
###############################################################################

# Logs specified string.
function logMsg() {
    if [ "$1" != "" ]; then
        echo "$(date): $1"
    else
        echo "$1"
    fi
}

# Prints the specified string if $DEBUG == 1.
function debugLog() {
    if (( $DEBUG == 1 )); then
        echo "$(date): $1"
    fi
}

# Prints a message to STDERR
function printError() {
    echo "$(date): $@" 1>&2
}

# Runs a command on a remote system via SSH and waits for the command to complete.
# Parameters:
# $1 - the SSH destination in the form "user@host".
# $2 - the command to be run remotely.
function sshRunCmd() {
    local destination="$1"
    local remoteCmd="$2"
    local port="${SSH_PORTS[$destination]}"
    local output=""

    debugLog "Running \"$remoteCmd\" on $destination"

    if [ "$port" != "" ]; then
        output=$(ssh -p "$port" "$destination" "$remoteCmd")
    else
        output=$(ssh "$destination" "$remoteCmd")
    fi

    debugLog "$output"
    RET="$output"
}

# Runs a command on all remote systems specified by the destinations array and
# waits for all remote systems to complete the command.
# Parameters:
# $1 - the name of the variable that contains the array of SSH destinations.
# $2 - the command to be run remotely.
function sshRunCmdOnMultipleSystems() {
    declare -n _destinations="$1"
    local remoteCmd="$2"
    debugLog "Running $remoteCmd on ${#_destinations[@]} remote systems."

    for _dest in "${_destinations[@]}"; do
        sshRunCmd "$_dest" "$remoteCmd" &
    done

    # Wait for all commands to complete.
    wait $(jobs -p)
}

# Busy waits until a certain operation is complete by calling a checkFn in a loop and sleeping between invocations.
# This function should only be used if an operation does not provide any better way of waiting, e.g., waiting for a process to exit.
# Parameters:
# $1 - the name of the checkFn to call. This function must set $RET to 0 to indicate that the operation is complete
#      and to any other value if the operation is not complete.
# $2 - the sleep duration.
function waitUntilDone() {
    local checkFn="$1"
    local sleepDuration="$2"

    eval "$checkFn"
    while (( "$RET" != 0 )); do
        echo "sleep $sleepDuration"
        sleep "$sleepDuration"
        eval "$checkFn"
    done
}
