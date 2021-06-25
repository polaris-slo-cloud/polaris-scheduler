# RAINBOW Orchestration

This repository contains the code for the orchestration work package (WP3) of the [RAINBOW](https://rainbow-h2020.eu/) project.

To clone this project, execute `git clone --recurse-submodules git@gitlab.com:rainbow-project1/rainbow-orchestration.git`

The code in the `ts` directory is taken from the [SLOC project](https://github.com/SLOCloud/SLOC), which is licensed under the Apache-2.0 License and which has been included in this repository as a submodule.

The submodule is hosted on GitHub and has been added with its `https` URL for compatibility reasons.
To access it via SSH, add the following to your `~/.gitconfig` file to always rewrite GitHub `https` URLs to SSH URLs:
```
[url "ssh://git@github.com/"]
    insteadOf = https://github.com/
```


## Documentation

The documentation for the RAINBOW orchestration components is available in the [docs](./docs) folder.

### Repository Organization

| Directory                | Contents |
|--------------------------|----------|
| [`deployment`](./deployment)         | YAML files to perform an all-in-one deployment of the RAINBOW orchestration stack |
| [`docs`](./docs)         | Documentation files (Work in progress) |
| [`go/orchestration`](./go/orchestration) | rainbow-orchestrator controller and CRDs (ServiceGraph, node topology) |
| [`go/scheduler`](./go/scheduler) | rainbow-scheduler: Fog-specific scheduler |
| [`ts`](./ts)             | TypeScript code (SLO controllers, etc.) |


## Deployment

To deploy the RAINBOW orchestration stack, open a terminal in the root folder of this repository and follow these steps:

1. Create the `rainbow-system` namespace:

```sh
kubectl create namespace rainbow-system
```

2. Create the `regcred` secret with your deployment token credentials for the [rainbow-integration](https://gitlab.com/rainbow-project1/rainbow-integration/container_registry) container registry:

```sh
kubectl create secret docker-registry -n=rainbow-system regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

3. Deploy the orchestrator components:

```sh
kubectl apply -f ./deployment
```

This deploys the following components:
* rainbow-scheduler
* Service Graph CRD
* Node Topology CRDs
* rainbow-orchestrator
* Horizontal Elasticity Strategy CRD and controller
* CRDs for the following SLOs:
    * **ToDo**
