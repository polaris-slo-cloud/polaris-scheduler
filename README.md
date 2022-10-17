# Polaris Scheduler

Polaris Scheduler is an SLO-aware Kubernetes scheduler.


## Documentation

The documentation for the polaris-scheduler is available in the [docs](./docs) folder.

### Repository Organization

| Directory                | Contents |
|--------------------------|----------|
| [`deployment`](./deployment)         | YAML files to perform an all-in-one deployment |
| [`docs`](./docs)         | Documentation files (Work in progress) |
| [`go/orchestration`](./go/orchestration) | CRDs (ServiceGraph, node topology) |
| [`go/scheduler`](./go/scheduler) | polaris-scheduler |



## Deployment

You must have a Kubernetes v1.21+ cluster available and configured in your KUBECONFIG file.
To deploy the polaris-scheduler, open a terminal in the root folder of this repository and execute the following:

```sh
kubectl apply -f ./deployment
```

This deploys the following components:
* polaris-scheduler
* Service Graph CRD
* Node Topology CRDs


## Experiments

This repository contains the following experiments for benchmarking Polaris Scheduler:

1. [Traffic Hazard Detection](./go/scheduler/hack/testbed/experiments/scenario-01/)
2. [Traffic Monitoring and Hazard Detection](./go/scheduler/hack/testbed/experiments/scenario-02/)


## Acknowledgement

Polaris-scheduler is a fork of the rainbow-scheduler and corresponding CRDs from the [orchestration repository](https://gitlab.com/rainbow-project1/rainbow-orchestration) of the [RAINBOW](https://rainbow-h2020.eu/) project.



[![DOI](https://zenodo.org/badge/449036242.svg)](https://zenodo.org/badge/latestdoi/449036242)
