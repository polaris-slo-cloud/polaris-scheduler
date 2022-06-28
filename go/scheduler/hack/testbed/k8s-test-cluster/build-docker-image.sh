#!/bin/sh

set -x
docker build . -t polarissloc/k8s-test-cluster:v0.0.2 -t polarissloc/k8s-test-cluster:latest
