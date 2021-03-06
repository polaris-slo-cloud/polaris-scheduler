# Polaris Scheduler

Polaris-Scheduler is an SLO-aware Kubernetes scheduler.


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

## Acknowledgement

Polaris-scheduler is a fork of the rainbow-scheduler and corresponding CRDs from the [orchestration repository](https://gitlab.com/rainbow-project1/rainbow-orchestration) of the [RAINBOW](https://rainbow-h2020.eu/) project.
