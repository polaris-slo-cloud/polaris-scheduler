#!/bin/bash
# set -x
set -o errexit
set -m

# This script uninstalls MicroK8s and must be run with sudo.

###############################################################################
# Global variables and imports
###############################################################################

SCRIPT_DIR=$(realpath $(dirname "${BASH_SOURCE}"))
source "$SCRIPT_DIR/../experiment.config.sh"
source "$SCRIPT_DIR/lib/util.sh"

# This script must be run with administrator privileges.
if [ "$SUDO_USER" == "" ]; then
    printError "This script must be run with sudo."
    exit 1
fi


###############################################################################
# Script Start
###############################################################################


# Uninstall MicroK8s
snap remove microk8s --purge

# Uninstall kubectl
snap remove kubectl --purge

echo "MicroK8s uninstallation complete."
