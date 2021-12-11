# RAINBOW Fog-aware Kubernetes Scheduler

## Node Roles

This scheduler assumes existence of the following node roles, i.e., `rainbow-h2020.eu/<role>` labels:

* `fog-node`: a fog node.
* `cloud-node`: a cloud node.

**ToDo**: Maybe we can relax this assumption to not require any fog or cloud labels?


## Application Components

This scheduler relies on each pod being associated with a ServiceGraph through the labels `rainbow-h2020.eu/service-graph` and `rainbow-h2020.eu/service-graph-node`.
Pods that do not have these labels will still be scheduled, but they cannot benefit from the optimizations brought by this scheduler.


## Scheduler Algorithm Overview

The RAINBOW scheduler extends the the default Kubernetes scheduler using the [scheduling-framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/).
It makes use of all the scheduling plugins enabled by [default](https://kubernetes.io/docs/reference/scheduling/config/#scheduling-plugins-1) and adds the following custom plugins:

| Plugin               | Extension Points      | Purpose |
|----------------------|-----------------------|---------|
| `ServiceGraph`       | `QueueSort`, `PreFilter`, `PostFilter`, `Reserve`, `Permit` | Load and cache the ServiceGraph of the pod's application, sort the pods, based on a breadth-first search on the ServiceGraph, and update (in-memory) the ServiceGraph with placement decisions. |
| `NetworkQoS`         | `PreFilter`, `Filter` | Filter out nodes that violate the network QoS constraints of the pod. |
| `NetworkQoS`         | `Score`, `NormalizeScore` | Prefer nodes that are likely to maintain the network QoS for a prolonged period of time. |
| `PodsPerNode`        | `PreScore`, `Score`, `NormalizeScore` | Increase colocation of an application's components on a node. |
| `NodeCost`           | `PreScore`, `Score`   | Give cheaper nodes a higher score. |
| `WorkloadType`       | `PreScore, `Score`, `NormalizeScore` | Prefer nodes that have worked well for the type of workload that the pod represents (boilerplate). |
| `AtomicDeployment`   | `Permit`              | Ensure that either all pods of an application are deployed or none of them. |


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
