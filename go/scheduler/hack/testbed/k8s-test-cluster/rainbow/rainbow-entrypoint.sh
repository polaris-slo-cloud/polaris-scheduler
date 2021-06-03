#!/bin/sh

ORIG_ENTRYPOINT="/usr/local/bin/entrypoint /sbin/init"

# Run the appropriate RAINBOW script
case "$K8S_NODE_TYPE" in
    "control-plane")
        /rainbow/control-plane-node/rainbow-control-node-setup.sh&
        ;;
    "worker")
        /rainbow/worker-node/rainbow-worker-node-setup.sh&
        ;;
    *)
        echo "Please set the K8S_NODE_TYPE environment variable either to 'control-plane' or 'worker' to define the type of this node."
        exit 1
esac

# The original entry point, which will start systemd needs to have PID 1, so we exec to it
exec $ORIG_ENTRYPOINT
