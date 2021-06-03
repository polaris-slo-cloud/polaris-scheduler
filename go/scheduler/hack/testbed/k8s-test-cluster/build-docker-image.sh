#!/bin/sh

set -x
docker build . -t rainbowh2020/k8s-test-cluster:v0.0.2 -t rainbowh2020/k8s-test-cluster:latest
