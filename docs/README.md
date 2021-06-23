# RAINBOW Orchestration Documentation

## Orchestration Components

The RAINBOW orchestration stack consists of the following components:

* **rainbow-scheduler**: fog-aware Kubernetes scheduler.
* **Service Graph**: a Kubernetes CRD to allow modelling a RAINBOW application with all its dependencies and configuration.
* **Node Topology Graph**: a Kubernetes CRD to model the RAINBOW cluster and the network link qualities among the nodes.
* **rainbow-orchestrator**: Kubernetes controller to handle the deployment and management of Service Graphs


## Service Graph

ToDo


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
