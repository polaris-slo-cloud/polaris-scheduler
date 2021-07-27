# RAINBOW Orchestration

This repository contains the code for the orchestration work package (WP3) of the [RAINBOW](https://rainbow-h2020.eu/) project.


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

You must have a Kubernetes v1.21+ cluster available and configured in your KUBECONFIG file.
To deploy the RAINBOW orchestration stack, open a terminal in the root folder of this repository and follow these steps:

1. Clone the `rainbow-orchestration` git repository:

    ```sh
    git clone git@gitlab.com:rainbow-project1/rainbow-orchestration.git
    ```


2. Create the `rainbow-system` namespace:

    ```sh
    kubectl create namespace rainbow-system
    ```


3. Create the `regcred` secret with your deployment token credentials for the [rainbow-integration](https://gitlab.com/rainbow-project1/rainbow-integration/container_registry) container registry:

    ```sh
    kubectl create secret docker-registry -n=rainbow-system regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
    ```


4. Deploy the orchestrator components:

    ```sh
    kubectl apply -f ./deployment
    ```

    This deploys the following components:
    * rainbow-scheduler
    * Service Graph CRD
    * Node Topology CRDs
    * rainbow-orchestrator
    * Horizontal Elasticity Strategy CRD and controller
    * CRDs and controllers for the following SLOs:
        * Image throughput SLO
