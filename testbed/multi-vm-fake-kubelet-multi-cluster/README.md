# Multi-VM Multi-Cluster Testbed

This testbed sets up a multi-cluster environment using multiple VMs.
Each VM will run one cluster with nodes simulated using fake-kubelet.


## Prerequisites

1. Clone the polaris-scheduler repository.

    ```sh
    git clone https://github.com/polaris-slo-cloud/polaris-scheduler.git
    ```

2. Install [MicroK8s](https://microk8s.io)

    ```sh
    cd polaris-scheduler/testbed/multi-vm-fake-kubelet-multi-cluster/prerequisites

    sudo ./1-setup-microk8s.sh
    ```
