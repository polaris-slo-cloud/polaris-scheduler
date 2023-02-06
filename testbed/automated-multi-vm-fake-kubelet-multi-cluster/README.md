# Multi-VM Multi-Cluster Testbed

This testbed sets up a multi-cluster environment, consisting of 10 clusters, using multiple VMs.
Each VM will run one cluster with nodes simulated using fake-kubelet.

By default, the 2-Smart Sampling mechanism is enabled, thus the following plugins are configured in the scheduler and the cluster-agents:

* [Polaris-Cluster-Agent](polaris-cluster-agent/1-config-map.yaml)
    * ResourcesFit (PreFilter, Filter, Score, CheckConflicts)
    * GeoLocation (PreFilter, Filter, Score)
    * BatteryLevel (PreFilter, Filter)
* [Polaris-Scheduler](./polaris-scheduler/polaris-scheduler-config.yaml)
    * RemoteNodesSampler

There is an alternative configuration (the `pure-random-sampling` subfolder of the [polaris-cluster-agent](./polaris-cluster-agent) and [polaris-scheduler](./polaris-scheduler) directories), which moves all but the CheckConflicts plugin from the cluster agent to the scheduler to simulate pure random sampling.
To use this configuration, please adapt the paths in the cluster configuration files for deploying the cluster agent and the path of the docker-compose file used to start the scheduler.


## Prerequisites

To run this testbed, the following VM configurations and pre-installed software are required:

* 10 VMs for simulating the clusters
    * Each VM should have at least 8 vCPUs and 32 GB of memory
    * [Snap](https://snapcraft.io) (for installing MicroK8s)
* 1 VM (or more) for running polaris-scheduler
    * 8 vCPUs and 32 GB of memory recommended
    * Docker
* 1 VM or device to serve as the load generator
    * [Apache JMeter](https://jmeter.apache.org)


### Set Up the Cluster VMs

This installation procedure assumes that `git` and `snap` are available on the VM.
Follow this procedure on all VMs that will host a cluster, i.e., all VMs that will run the polaris-cluster-agent.

1. Clone the polaris-scheduler repository.

    ```sh
    git clone https://github.com/polaris-slo-cloud/polaris-scheduler.git
    ```

2. Install [MicroK8s](https://microk8s.io)

    ```sh
    cd polaris-scheduler/testbed/multi-vm-fake-kubelet-multi-cluster

    # Optional: use a specific MicroK8s version
    export MICRO_K8S_CHANNEL="1.25/stable"

    # Install MicroK8s
    sudo ./prerequisites/1-setup-microk8s.sh
    ```


### Start a Cluster

To start a cluster on the current VM, please use the [start-cluster.sh](./start-cluster.sh).
It requires the path to the cluster config file to be either passed as an argument or be set as the `POLARIS_TESTBED_CONFIG` environment variable.
Using the environment variable is recommended, because it allows you to easily "start" (deploy the polaris and fake-kubelet components) and "stop" (undeploy the components) the cluster easily.

The [clusters](./clusters) folder contains configuration files for cluster groups consisting of 1k, 5k, 10k, 15k, and 20k nodes total.
Each cluster group folder contains 10 configuration files for the individual clusters.
Each cluster VM needs to use its own configuration file, i.e., VM1 uses `cluster-01.config.sh`, VM2 uses `cluster-02.config.sh`, etc.
These configuration files can be adapted as needed.

To start a cluster on a VM, do the following:

1. Set `POLARIS_TESTBED_CONFIG` to point to this configuration file.

    ```sh
    # Example
    export POLARIS_TESTBED_CONFIG=./clusters/20k-nodes/cluster-01.config.sh
    ```

2. Start the testbed components.

    ```sh
    ./start-cluster.sh
    ```

3. Wait for all fake-kubelet pods to be running.

    ```sh
    # Check for non-ready nodes.
    kubectl get nodes | grep --invert-match Ready

    # Check if any pods are not running.
    kubectl get pods -A | grep --invert-match Running

    # Check that the polaris-cluster-agent is running.
    kubectl get pods -n polaris
    ```

4. Create a directory for storing logs from the experiment runs.

    ```sh
    mkdir results
    ```

After you have set up the clusters on all VMs, you can proceed to setting up the polaris-scheduler and subsequently running the experiments.

**Note** that to remove the fake-kubelet nodes again, you need to uninstall MicroK8s on the machine.


### Configure and Start Polaris-Scheduler

To start the polaris-scheduler, follow this procedure on as many VMs as you need instances:

1. Clone the polaris-scheduler repository and navigate to the directory with the polaris-scheduler docker-compose configuration.

    ```sh
    git clone https://github.com/polaris-slo-cloud/polaris-scheduler.git
    cd polaris-scheduler/testbed/multi-vm-fake-kubelet-multi-cluster/polaris-scheduler
    ```

2. Configure the addresses of the clusters in the [polaris-scheduler-config.yaml](./polaris-scheduler/polaris-scheduler-config.yaml) config file.

3. Start the scheduler. By default, it will be accessible on port `38080`.

    ```sh
    docker compose up -d
    ```

4. Create a directory for storing logs from the experiment runs.

    ```sh
    mkdir results
    ```


## Executing Commands on Multiple VMs

Saving logs on and resetting the cluster VMs after an experiment run, requires running commands on all VMs.
To ease this process, we recommend using either `pssh` (parallel ssh) or `tmux` with pane synchronization enabled.

To use `tmux`, open a new `tmux` window and 10 panes inside.
Connect to one VM within each pane.
Enable pane synchronization using:

    ```sh
    # Press the key combination
    Ctrl + b + :

    # Execute the following command
    set synchronize-panes on
    ```

Every command that you type or paste will now go to all panes.


Alternatively, to use `pssh`, create a file `test-hosts.txt` with the list of all `user@VM` ssh connection strings, e.g.,

    ```
    user@192.168.0.10
    user@192.168.0.11
    ...
    ```

You can then execute commands on all these VMs by running:

    ```sh
    # On Ubuntu pssh is called parallel-shh
    parallel-ssh -h ./test-hosts.txt -i "<Command>"
    ```


## Run the Experiments

This repository includes multiple [test plans](./jmeter-test-plans)) designed with Apache JMeter.
Feel free to adapt them to your needs and then run some experiments.
All pods in this experiment are created in the `test` namespace.

To run an experiment, execute the following command on your load generator VM:

    ```sh
    jmeter.sh -n -t <test-plan>

    # Example:
    jmeter.sh -n -t ./jmeter-test-plans/heterogeneous-pods/polaris-scheduler-10ms-5threads.jmx
    ```


### Export Logs

For simplicity, we assume that we want to save the logs of the current experiment to the file `results/01.log` on every VM (adapt the name as needed for each experiment).

1. Make sure that all pods have completed scheduling (or have failed).
To this end, display the scheduler container logs and ensure that no new lines are being added.

    ```sh
    # On the polaris-scheduler VM
    docker compose logs -f --tail=100 polaris-scheduler
    ```

2. Export the scheduler log.

    ```sh
    # On the polaris-scheduler VM
    docker compose logs polaris-scheduler > results/01.log
    ```

3. Export the logs of all polaris-cluster agents.

    ```sh
    # Option 1: Using tmux on your local computer (paste the command into all panes):
    kubectl logs -n polaris `kubectl get pods -n polaris -o=custom-columns="name:.metadata.name" | grep polaris` > ./results/01.log

    # Option 2: Using pssh on your local computer:
    parallel-ssh -h ./test-hosts.txt -i "kubectl logs -n polaris \`kubectl get pods -n polaris -o=custom-columns=\"name:.metadata.name\" | grep polaris\` > \$HOME/polaris-scheduler/testbed/multi-vm-fake-kubelet-multi-cluster/results/01.log"
    ```


### Reset Clusters after an Experiment

To reset the clusters for a new experiment, you need to do the following on every cluster: scale the polaris-cluster-agent to zero replicas (to clear the log), delete the `test` namespace (and all its pods), recreate the `test` namespace, and scale the polaris-cluster-agent back to 1 replica.
Use one of the following all-in-one commands:

    ```sh
    # Option 1: Using tmux on your local computer (paste the command into all panes).
    kubectl scale deployment -n polaris polaris-cluster-agent --replicas=0 && kubectl delete namespace test && kubectl create namespace test && kubectl scale deployment -n polaris polaris-cluster-agent --replicas=1 && kubectl get pods -n polaris

    # Option 2: Using pssh on your local computer.
    parallel-ssh -h ./gcp-hosts.txt -i "kubectl scale deployment -n polaris polaris-cluster-agent --replicas=0 && kubectl delete namespace test && kubectl create namespace test && kubectl scale deployment -n polaris polaris-cluster-agent --replicas=1 && kubectl get pods -n polaris"
    ```

If the deletion of the `test` namespace times out, it can help to wait a few minutes for the pods to be terminated.
Alternatively, you can force delete pods, which are stuck of reinstall MicroK8s.

Additionally, restart the polaris-scheduler:

    ```sh
    docker compose down
    docker compose up -d
    ```


## Miscellaneous Useful Commands

```sh
# Get name of cluster agent pod for integration in another command
kubectl get pods -n polaris -o=custom-columns="name:.metadata.name" | grep polaris
# Example:
kubectl logs -n polaris `kubectl get pods -n polaris -o=custom-columns="name:.metadata.name" | grep polaris`

# Check the status of the cluster agents using pssh
parallel-ssh -h ./test-hosts.txt -i  "curl localhost:30033/samples/status"

# Check if any pods are not running using pssh
parallel-ssh -h ./test-hosts.txt -i "kubectl get pods -A | grep --invert-match Running"

# Check for non-ready nodes using pssh
parallel-ssh -h ./test-hosts.txt -i "kubectl get nodes | grep --invert-match Ready"

# Force delete all pods that are stuck in Terminating and then recreate the test namespace
kubectl get pods -A | grep Terminating | awk '{print $2 " --namespace=" $1}' | xargs kubectl delete pod --force && sleep 10s && kubectl delete namespace test && kubectl delete namespace test && kubectl create namespace test

# Quickly add all VMs to your known_hosts file (without checking)
parallel-ssh -O StrictHostKeyChecking=no -h ./test-hosts.txt -i hostname
```
