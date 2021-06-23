# RAINBOW Kubernetes Controllers

This directory contains all RAINBOW's Kubernetes controllers and CRDs written in Go, except for the scheduler.

There are some RAINBOW-specific data structures that are used in the controllers (see [docs](../../docs) for further details):

* Service Graph
* Node Topology Graph


## Testbed

The testbed found in [hack/kind-cluster](./hack/kind-cluster) uses [kind](https://kind.sigs.k8s.io) to create a local Kubernetes cluster in Docker containers.
Any other Kubernetes cluster may be used as well - it only needs to be set as the current context in `kubectl`.

To run the RAINBOW controller in the testbed, execute the following steps:

1. Open a terminal in the folder [hack/kind-cluster](./hack/kind-cluster) and run
    ```sh
    # Creates a test cluster consisting of 4 nodes and a private Docker registry.
    ./start-kind-cluster.sh
    ```

1. To build the controller, open a terminal in the `go` root directory of this repository and run
    ```sh
    make
    ```

1. Apply all CRDs by running
    ```sh
    make install
    ```

1. Run the controller locally:
    ```sh
    cd bin
    ./manager
    ```

1. Optionally, apply the sample resources to the cluster by opening a terminal in the `go` directory and running
    ```sh
    kubectl apply -f ./config/samples
    ```


## Directory Structure

| Directory                | Contents |
|--------------------------|----------|
| [`apis`](./apis)         | Go types for the CRDs. |
| [`config`](./config)     | Deployment manifests, CRDs, configuration, and [example files](./config/samples). |
| [`controllers`](./controllers)| Kubernetes controllers for the CRDs. |
| [`hack`](./hack)         | Scripts and code generation templates. |
| [`internal`](./internal) | Types and functions for internal use only. |
| [`pkg`](./pkg)           | Types and functions that can be reused by other projects. |
