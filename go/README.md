# RAINBOW Kubernetes Controllers

This directory contains RAINBOW's Kubernetes controllers and CRDs written in Go.

## Node Topology Graph

RAINBOW needs to know the node topology graph of the cluster.
For this graph, the set of nodes and links are stored separately:

* The set of nodes is the standard list of Nodes obtainable through the Kubernetes API.
* The set of links uses the RAINBOW `NetworkLink` CRD.

Both lists together can be used to construct a node topology graph.
The advantage of not storing one big node topology graph CRD is that the graph's elements can be updated separately whenever new metrics are available.

Each `NetworkLink` connects two nodes that have a direct network connection to each other.
In the CRD there are two fields `nodeA` and `nodeB`, which refer to the names of these nodes.
Each link `nodeA <-> nodeB` exists only once - the admission webhook generates the name of a link automatically by sorting the two node names alphabetically and concatenating them, thus, ensuring that a link with `nodeA` and `nodeB` swapped is not stored as a duplicate. 


## Testbed

The testbed found in [hack/kind-cluster-with-local-registry](./hack/kind-cluster-with-local-registry) uses [kind](https://kind.sigs.k8s.io) to create a local Kubernetes cluster in Docker containers.
Any other Kubernetes cluster may be used as well - it only needs to be set as the current context in `kubectl`.

To run the RAINBOW controller in the testbed, execute the following steps:

1. Open a terminal in the folder [hack/kind-cluster-with-local-registry](./hack/kind-cluster-with-local-registry) and run
    ```sh
    # Creates a test cluster consisting of 4 nodes and a private Docker registry.
    ./start-kind-with-local-registry.sh
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
