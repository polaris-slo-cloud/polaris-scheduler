# Kubernetes Test Cluster

This directory contains the files needed for creating a Docker image that can be used to bootstrap a Kubernetes test cluster, where each node is represented by a Docker container.
The purpose of this project is to allow creating Kubernetes clusters in the [Fogify](https://ucy-linc-lab.github.io/fogify) simulator, which requires docker-compose files as input.

This is based on the Docker image and configuration from the [kind](https://kind.sigs.k8s.io/) project.
We would like to thank the `kind` team for their great work in creating a Kubernetes distribution that is easily executable within Docker.


## Project Layout

This project consists of the following main parts:

* `rainbow/`: Contains the scripts and configuration files used to bootstrap Kubernetes in a Docker container. The configuration files were taken from the `kind` project.
* `data/`: Used for mounting a host volume in the containers when started with docker-compose.
* `fogify-k8s/`: Contains a docker-compose file for starting a sample Kubernetes cluster in Fogify.
* `docker-compose.yaml`: Used for running the k8s-test-cluster without building a Docker image first.
* `Dockerfile` and `build-docker-image.sh`: Used for building a k8s-test-cluster Docker image for use in Fogify.


## Running k8s-test-cluster

### Running Directly in docker-compose

To run a k8s-test-cluster directly using docker-compose:

1. Open a terminal in the `k8s-test-cluster` root directory
1. Run `docker-compose up -d` to start the cluster
1. You can use `docker-compose logs -f` to check on the setup progress
1. To interact with the cluster you need to use the KUBECONFIG file created by the control plane node (`data/kubeconfig/config.public`).
    * The easiest way is to set the $KUBECONFIG environment variable and to then use `kubectl` normally, e.g., `export KUBECONFIG=$(pwd)/data/kubeconfig/config.public`

### Running in Fogify

1. Open a terminal in the `k8s-test-cluster` root directory.
1. Build the Docker image by executing `./build-docker-image.sh`
1. Set `CONNECTOR=DockerComposeConnector` in your `.env` file, as described [here](https://github.com/UCY-LINC-LAB/fogify/tree/master/examples/kind-project).
1. Use Fogify to start the cluster using a docker-compose file based on `fogify-k8s/docker-compose.yaml`.


## How this Docker Image was Developed

We created a test cluster using `kind` and then examined the resulting Docker containers, as well as the `kind` source code.
We rebuilt the steps that `kind` executes using some shell scripts (in a simplified fashion).

The following Sections explain our analysis of `kind` in detail.

### `kind create` Configuration

The configuration for this test cluster was obtained by analyzing the `kind create` command with the following input configuration:

```YAML
# three node (two workers) cluster config
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
```

The command `kind create cluster --config ./kind-cluster-config.yml` produced the following output:
```
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.19.1) 🖼
 ✓ Preparing nodes 📦 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Thanks for using kind! 😊
```


### `kind create` Code Analysis

The analysis of the [kind create](https://github.com/kubernetes-sigs/kind/tree/master/pkg/cmd/kind/create) source code yields the following initialization steps:

1. Create a distinct Docker network for the cluster.
    * `docker network create -d=bridge -o com.docker.network.bridge.enable_ip_masquerade=true kind`
1. Provision nodes (=containers) using `docker run`.
    * Image: `kindest/node:v1.19.4`
    * `docker run` arguments (see [provision.go](https://github.com/kubernetes-sigs/kind/blob/master/pkg/cluster/internal/providers/docker/provision.go) in `kind` src):
        ```Go
        // From runArgsForNode()
        "--hostname", name, // make hostname match container name
        "--name", name, // ... and set the container name
        // label the node with the role ID
        "--label", fmt.Sprintf("%s=%s", nodeRoleLabelKey, node.Role),
        // running containers in a container requires privileged
        // NOTE: we could try to replicate this with --cap-add, and use less
        // privileges, but this flag also changes some mounts that are necessary
        // including some ones docker would otherwise do by default.
        // for now this is what we want. in the future we may revisit this.
        "--privileged",
        "--security-opt", "seccomp=unconfined", // also ignore seccomp
        "--security-opt", "apparmor=unconfined", // also ignore apparmor
        // runtime temporary storage
        "--tmpfs", "/tmp", // various things depend on working /tmp
        "--tmpfs", "/run", // systemd wants a writable /run
        // runtime persistent storage
        // this ensures that E.G. pods, logs etc. are not on the container
        // filesystem, which is not only better for performance, but allows
        // running kind in kind for "party tricks"
        // (please don't depend on doing this though!)
        "--volume", "/var",
        // some k8s things want to read /lib/modules
        "--volume", "/lib/modules:/lib/modules:ro",

        // From commonArgs()
        "--detach", // run the container detached
        "--tty",    // allocate a tty for entrypoint logs
        // label the node with the cluster ID
        "--label", fmt.Sprintf("%s=%s", clusterLabelKey, cluster),
        // user a user defined docker network so we get embedded DNS
        "--net", networkName,
        // [...]
        // What we desire is:
        // - restart on host / dockerd reboot
        // - don't restart for any other reason
        // [...]
        // so the closest thing is on-failure:1, which will retry *once*
        "--restart=on-failure:1",
        ```
    * The image's entry point is `ENTRYPOINT [ "/usr/local/bin/entrypoint", "/sbin/init" ]`.
        * `/usr/local/bin/entrypoint` is a shell script, which sets some things up and then executes the argument given to the script (`/sbin/init`).
        * `/sbin/init` points to `systemd`, which must be executed with PID 1.
1. Set up external load balancer (not executed by default).
1. Set up `kubeadm` configuration on control plane nodes.
    * See [kubeadm.conf](./control-plane-node/kind/kubeadm.conf)  (this file is generated by kind)
1. `kubeadm init` on primary control plane node.
    * `kubeadm init --skip-phases=preflight --config=/kind/kubeadm.conf --skip-token-print --v=6`
    * Copy some files to secondary control plane nodes (we will skip this for now).
1. Install CNI on primary control plane node.
    * Generate a configuration (see [output](./control-plane-node/kind/manifests/default-cni.yaml)) and pipe it into `kubectl create`
    * Equivalent: `kubectl create --kubeconfig=/etc/kubernetes/admin.conf -f /kind/manifests/default-cni.yaml`
1. Install StorageClass on primary control plane node.
    * `kubectl --kubeconfig=/etc/kubernetes/admin.conf apply -f /kind/manifests/default-storage.yaml`
1. `kubeadm join` on secondary control plane nodes and worker nodes.
    * `kubeadm join --config /kind/kubeadm.conf --skip-phases=preflight --v=6`
1. Wait until the cluster is ready.
    * `kubectl --kubeconfig=/etc/kubernetes/admin.conf get nodes --selector=node-role.kubernetes.io/master -o=jsonpath='{.items..status.conditions[-1:].status}'`
1. Adapt user's `kubeconfig` to allow access via `kubectl`.
