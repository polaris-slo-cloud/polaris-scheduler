# Multi-VM Multi-Cluster Testbed

This testbed sets up a multi-cluster environment using multiple VMs.
Each VM will run one cluster with nodes simulated using fake-kubelet.


## Prerequisites

This installation procedure assumes that `git` and `snap` are available on the VM.

1. Clone the polaris-scheduler repository.

    ```sh
    git clone https://github.com/polaris-slo-cloud/polaris-scheduler.git
    ```

2. Install [MicroK8s](https://microk8s.io)

    ```sh
    cd polaris-scheduler/testbed/multi-vm-fake-kubelet-multi-cluster

    sudo ./prerequisites/1-setup-microk8s.sh
    ```


## Start a Cluster

To start a cluster on the current VM, please use the [start-cluster.sh](./start-cluster.sh).
It requires the path to the cluster config file to be either passed as an argument or be set as the `POLARIS_TESTBED_CONFIG` environment variable.
Using the environment variable is recommended, because it allows you to easily "start" (deploy the polaris and fake-kubelet components) and "stop" (undeploy the components) the cluster easily.

1. Adapt one of the cluster configuration files in the [clusters](./clusters) folder or create a new one.

2. Set `POLARIS_TESTBED_CONFIG` to point to this configuration file.

    ```sh
    # Example
    export POLARIS_TESTBED_CONFIG=./clusters/cluster-01.config.sh
    ```

3. Start the testbed components.

    ```sh
    ./start-cluster.sh
    ```

4. Wait for all fake-kubelet pods to be running.

    ```sh
    # Check if any pods are still pending.
    kubectl get pods -n fake-kubelet -o wide | grep Pending
    ```

After you have set up the clusters on all VMs, you can proceed to setting up the polaris-scheduler and subsequently running the experiments.

Note that to remove the fake-kubelet nodes again, you need to uninstall MicroK8s on the machine.
