# Polaris Scheduler Workspace Organization

This directory contains the modules that comprise the Polaris Scheduler workspace:

* [framework](./framework): The orchestrator-independent Polaris Scheduler framework library that defines the scheduling pipeline and plugin structures.
* [k8s-connector](./k8s-connector): Kubernetes orchestrator connector.
* [cluster-agent](./cluster-agent): The polaris-cluster-agent executable module that relies on the k8s-connector module.
* [scheduler](./scheduler): The polaris-scheduler executable module.
* [context-awareness](./context-awareness): Context-aware scheduling plugins
