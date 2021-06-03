# RAINBOW Kubernetes Controllers

This directory contains RAINBOW's Kubernetes controllers and CRDs written in Go, which can be found in the subfolders of this directory:

* [orchestrator](./orchestration): All shared CRDs and data structures and the controller for the RAINBOW CRDs.
* [scheduler](./scheduler): Fog-aware Kubernetes scheduler, based on kube-scheduler.
