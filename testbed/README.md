# Testbeds

This folder contains scripts and configuration files to set up various testbeds for the Polaris Scheduler.

| Folder | Testbed Configuration |
|--------|-----------------------|
| [fake-kubelet-cluster](./fake-kubelet-cluster) | A single local [kind](https://kind.sigs.k8s.io) cluster with simulated nodes. |
| [fake-kubelet-multi-cluster](./fake-kubelet-multi-cluster) | 10 local [kind](https://kind.sigs.k8s.io) clusters with simulated nodes. |
| [multi-vm-fake-kubelet-multi-cluster](./multi-vm-fake-kubelet-multi-cluster) | 10 [MicroK8s](https://microk8s.io) clusters, each on its own VM, with 1k-20k simulated nodes. |
