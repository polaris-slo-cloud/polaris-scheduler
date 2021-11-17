# RAINBOW Fog-aware Kubernetes Scheduler

## Node Roles

This scheduler assumes existence of the following node roles, i.e., `node-role.kubernetes.io/<role>` labels:

* `control-plane`
* `master`
* `worker`: a worker node (can be in the fog or in the cloud).
* `fog-region-head`: the cluster head of a fog region.
* `fog`: a fog node.
* `cloud`: a cloud node.

**ToDo**: Maybe we can relax this assumption to not require any fog or cloud labels?


## Application Components

Currently this scheduler is designed for asynchronous applications.
It is necessary to annotate the pods of the message queue with the label: `app.kubernetes.io/component: message-queue`.


## Scheduler Algorithm Overview

The RAINBOW scheduler extends the the default Kubernetes scheduler using the [scheduling-framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/).
It makes use of all the scheduling plugins enabled by [default](https://kubernetes.io/docs/reference/scheduling/config/#scheduling-plugins-1) and adds the following custom plugins:

| Plugin                      | Extension Points    | Purpose |
|-----------------------------|---------------------|---------|
| `RainbowPriorityMqSort`     | `QueueSort`         | Ensure that a message queue is scheduled before other pods. |
| `RainbowServiceGraph`       | `PreFilter`         | Get and cache the service graph of the application, to which the pod belongs. |
| `RainbowLatency`            | `Filter`            | Filter out nodes that violate the latency constraints of the application. |
| `RainbowPodsPerNode`        | `PreScore`, `Score`, `NormalizeScore` | Increase colocation of an application's components on a node. |
| `RainbowNodeCost`           | `PreScore`, `Score` | Give cheaper nodes a higher score. |
| `RainbowReserve`            | `Reserve`           | Updates the service graph with the info about the selected node. |
| `RainbowAtomicDeployment`   | `Permit`            | Ensure that either all pods of an application are deployed or none of them. |


**ToDo:**
- Move removal of cloud nodes if there is at least one eligible fog node from  RainbowPodsPerNode to a PostFilter plugin
- Merge RainbowReserve into the RainbowServiceGraph plugin

## Some notes about the scheduler extension points

Info taken from https://www.youtube.com/watch?v=Wr1TMbdc4O0

![Scheduling Framework Extension Points](https://d33wubrfki0l68.cloudfront.net/4e9fa4651df31b7810c851b142c793776509e046/61a36/images/docs/scheduling-framework-extensions.png)

*This image is licensed under CC BY 4.0 and was taken from https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/ .

* **PreFilter:** useful for doing cross-node checks and cache some info for the `Filter` phase.
* **PostFilter:** can be used for preemption
* **Permit:** useful, e.g., when scheduling a batch application, to wait until all components have been assigned to a pod and then permit all of them at once (can be used to simulate an atomic deployment operation).


## Test Scenarios for Comparing Against Other Schedulers

We have prepared a set of scenarios to compare rainbow-scheduler against other scheduling approaches.
These scenarios can be found in [this repository](https://gitlab.com/tommazzo89/scheduler-test-scenarios).


## Building and Debugging

### Building

To build the scheduler run the following command:

```sh
make build
```

To build the Docker image of the scheduler run the following command:

```sh
make release-image
```


### Debugging

To debug the scheduler with VS Code, please follow these steps:

1. Copy the `/etc/kubernetes/scheduler.conf` file from your Kubernetes cluster's master node to `bin/config/kubernetes/scheduler.conf`.

2. Go to the "Run and Debug" view in VS Code, select "Local: Debug Scheduler" from the dropdown list, and click the "Start Debugging" button.
This will automatically execute `make build-debug` and launch the scheduler with the required command line parameters under the Go debugger.
