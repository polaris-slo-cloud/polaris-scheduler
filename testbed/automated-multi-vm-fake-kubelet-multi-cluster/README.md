# Automated Multi-VM Multi-Cluster Testbed

This testbed contains scripts to automatically set up a multi-cluster environment and run experiments on it.
The environment consists of 10 clusters, using multiple VMs.
Each VM runs one cluster with nodes simulated using fake-kubelet.

Experiment execution is completely automatic:
setup of the Kubernetes clusters and polaris-scheduler, experiment execution, and results collection are automatically done by the [run-experiments.sh](./run-experiments.sh) script.
The only manual steps required are the provisioning of the 12 VMs, cloning of the git repository, and providing the configuration.

The whole experiment suite consists of multiple configurations with varying polaris-scheduler and polaris-cluster-agent settings, as well as different cluster sizes.
The following table provides a quick overview of the contents of this testbed.

| Directory/File           | Contents |
|--------------------------|----------|
| [`clusters`](./clusters) | Cluster configurations for the 10 subclusters, for total supercluster sizes of 1k, 5k, 10k, 15k, and 20k nodes. |
| [`experiments`](./experiments) | Experiment suites that consist of multiple experiments each. See [Experiment Suites](#experiment-suites) for details. |
| [`jmeter-test-plans`](./jmeter-test-plans) | [Apache JMeter](https://jmeter.apache.org) test plans for generating load during the experiments. |
| [`polaris-cluster-agent`](./polaris-cluster-agent) | Configurations for the polaris-cluster-agent. Each subdirectory contains a distinct configurations. |
| [`polaris-scheduler`](./polaris-scheduler) | Configurations for the polaris-scheduler. Each subdirectory contains one or more distinct configurations. |
| [`scripts`](./scripts) | The scripts used during the experiments. |
| [`experiment.config.sh`](./experiment.config.sh) | The global experiment configuration. This file contains, e.g., the addresses of the VMs and references the distinct experiment files from the [`experiments`](./experiments) that should be executed. |
| [`extract-scheduler-results.sh`](./extract-scheduler-results.sh) | Script to extract results from the experiments' log files and store them in a CSV file. |
| [`run-experiments.sh`](./run-experiments.sh) | The main script to run the experiments. |


## Prerequisites

To run this testbed, the following VM configurations and pre-installed software are required:

* 10 VMs for simulating the clusters
    * Each VM should have at least 8 vCPUs and 32 GB of memory
    * [Snap](https://snapcraft.io) (for installing MicroK8s)
    * Password-less sudo for the user that is used to run the experiments (needed for automatically installing [MicroK8s](https://microk8s.io))
* 1 VM (or more) for running polaris-scheduler
    * 8 vCPUs and 32 GB of memory recommended
    * Docker
* 1 VM or device to serve as the experiment coordinator and load generator
    * [Apache JMeter](https://jmeter.apache.org)
    * SSH access using public key authentication to all the other VMs


## Running the Experiments

1. Clone the polaris-scheduler repository on all VMs.
The local path of the repository must be the same on all cluster VMs
See [Executing Commands on Multiple VMs](#executing-commands-on-multiple-vms) for hints on how to easily run commands on all VMs.

    ```sh
    git clone https://github.com/polaris-slo-cloud/polaris-scheduler.git
    ```

2. On the experiment coordinator/load generator VM, navigate to the `polaris-scheduler/testbed/automated-multi-vm-fake-kubelet-multi-cluster` directory.

3. Open the [experiment.config.sh](./experiment.config.sh), which contains the configuration for the experiments and set the following variables (detailed explanations are given in the config file):

    * `JMETER_SH`
    * `POLARIS_SCHED_REPO_NODE_VMS`
    * `POLARIS_SCHED_REPO_SCHEDULER_VM`
    * `CLUSTER_VMS`
    * `SCHEDULER_VM`

4. If you wish to modify the list of experiments that will be executed, adapt the `EXPERIMENT_ITERATIONS` associative array accordingly. See [Experiment Suites](#experiment-suites) for a description of the available experiments.

5. To execute the experiments (and log the console output), run the following command:

    ```sh
    ./run-experiments.sh 2>&1 | tee experiments.log
    ```

6. Each experiment iteration involves the installation of a MicroK8s cluster and the Polaris Scheduler components on every cluster VM, execution of the experiment, waiting for all pods to be processed, and teardown of the MicroK8s clusters. Thus, the entire default set of experiments can take multiple hours.

7. After all experiments are completed, the results (including logs from the cluster VMs) are available in the director configured in the `RESULTS_ROOT` variable (by default `./results`).



## Experiment Suites

The [experiments](./experiments) folder contains multiple experiment suites, each in its own subfolder.
Each suite contains a set of experiment iterations (individual `.sh` files in the subfolders), each with a particular configuration.
Each experiment references a cluster configuration from the [clusters](./clusters) folder.
This folder contains configurations for the 10 subclusters, for total supercluster sizes of 1k, 5k, 10k, 15k, and 20k nodes.

The following subsections briefly describe the currently available experiment suites.


### 2-smart-sampling-tuning

The [2-smart-sampling-tuning](./experiments/2-smart-sampling-tuning) suite can be used to tune two important configuration values for 2-smart-sampling: `percentageOfClustersToSample` of the `RemoteNodesSampler` plugin and `nodesToSampleBp`, i.e., the percentage of clusters to sample for each pod and the percentage of nodes to sample within each cluster.

This experiment uses the 20k nodes supercluster and creates 11,200 pods with 4 CPUs and 4 GiB RAM, which is the maximum number of pods of this size that the supercluster can support (note that 50% of the nodes are intentionally too small to host such a pod).
Thus, scheduling all these pods entails a perfect placement that does not miss any eligible node that still has free resources.

The experiment suite runs through `percentageOfClustersToSample` settings of {10%, 20%, ..., 100%} and, for each of them, runs four experiment iterations, one for each `nodesToSampleBp` setting in the set of {400, 800, 1200, 1600} (i.e., sampling 4%, 8%, 12%, and 16% of the nodes per cluster).


### 2-smart-sampling-tuning-lru-node

The [2-smart-sampling-tuning-lru-node](./experiments/2-smart-sampling-tuning-lru-node) suite is the same as [2-smart-sampling-tuning](#2-smart-sampling-tuning), except that is uses the `LeastRecentlyUsedNode` plugin in the scoring phase of the Cluster Agent to assign node scores that are proportional to how long a node has not been assigned a new pod (nodes that have not received a pod for a long time, get the highest scores - least recently used).


### scalability-nodes

The [scalability-nodes](./experiments/scalability-nodes/) suite aims to assess the scalability of Polaris Scheduler on increasing cluster sizes (1k, 5k, 10k, 15k, and 20k total nodes in the supercluster).
Each experiment iteration submits 1000 pods to the scheduler, varying the sizes of the clusters across the iterations.
These pods intentionally fit on every node.
The aim of this experiment suite is to compare the execution times of various scheduling phases across the different cluster sizes.


### scalability-jobs

The [scalability-jobs](./experiments/scalability-jobs) suite contains 15 iterations that gradually increase the number of pods (jobs) that are submitted to the scheduler per second.
Since JMeter does not allow for configuring a specific request rate per second, but instead requires configuring the number threads that generate requests and the approximate timing they should use, the final request rate may vary across different systems, depending on how long it takes a request to be executed.
The supercluster used for this experiment contains 20k nodes in total.
This experiment suite is intended to serve as a stress test for the scheduler.


### scalability-jobs-pure-random

The [scalability-jobs-pure-random](./experiments/scalability-jobs-pure-random) is like the [scalability-jobs](#scalability-jobs) suite, except that it uses pure random sampling, i.e., no 2-smart-sampling.



## Results Processing

### Statistics

The [extract-scheduler-results.sh](./extract-scheduler-results.sh) script can be used to process the experiment results and extract various aggregated metrics from them.
To use it, follow these steps:

1. Copy the results directory from the load generator VM.

2. Execute the `extract-scheduler-results.sh`, specifying the results directory as the first argument and the output CSV file as the second.

    ```sh
    # Usage:
    # ./extract-scheduler-results.sh <results directory> <csv output file>
    ./extract-scheduler-results.sh ./results ./output.csv
    ```


### Individual Metrics

In certain cases it may be necessary to extract specific metrics in a non-aggregated form.
To this end, we provide a set of ready-to-use commands with regular expressions below.

All of the following commands operate on the log files of the scheduler and require extracting the successful scheduler runs first.
To extract the successful runs, please execute the following command in the directory that contains the scheduler logs (by default `results/scheduler`):

```sh
mkdir successes

# Extract successful runs and store them in the successes directory.
for f in ./*.log; do cat "$f" | grep SchedulingSuccess > "./successes/$f-success.log"; done
```

In the `successes` directory, you can use the following commands to extract individual metrics and export them to distinct files.

```sh
# Extract samplingDuration:
for f in ./*-success.log; do cat "$f" | awk '{match($0, /^.+samplingDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' > "./samplingDuration/$f"; done

# Extract e2eDuration:
for f in ./*-sampling.log; do cat "$f" | awk '{match($0, /^.+e2eDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' > "./e2eDuration/$f"; done

# Extract commitDuration:
for f in ./*-success.log; do cat "$f" | awk '{match($0, /^.+commitDurationMs"=([0-9]+)\s.+/, arr); print arr[1];}' > "./commitDuration/$f"; done

# Extract queueTime:
for f in ./*-success.log; do cat "$f" | awk '{match($0, /^.+queueTimeMs"=([0-9]+)\s.+/, arr); print arr[1];}' > "./queueTime/$f"; done

# Extract the number of sampledNodes:
for f in ./*-success.log; do cat "$f" | awk '{match($0, /^.+sampledNodes"=([0-9]+)\s.+/, arr); print arr[1];}' > "./sampledNodes/$f"; done
```


## Executing Commands on Multiple VMs

Initial setup and deleting the logs after a set of experiments, requires running commands on all VMs.
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

```sh
user@192.168.0.10
user@192.168.0.11
...
```

You can then execute commands on all these VMs by running:

```sh
# On Ubuntu pssh is called parallel-shh
parallel-ssh -h ./test-hosts.txt -i "<Command>"
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
parallel-ssh -O StrictHostKeyChecking=no -h ./vms.txt -i hostname
```

