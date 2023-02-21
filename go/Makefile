# Copyright 2022 Polaris SLO Cloud Project.
# Based on the Makefile from https://github.com/kubernetes-sigs/scheduler-plugins
# created by the Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

COMMONENVVAR=GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m)))
BUILDENVVAR=CGO_ENABLED=0

PROJECT_VERSION=2.0.0-beta.32

# RELEASE_REGISTRY is the container registry to push into.
RELEASE_REGISTRY?=polarissloc
# RELEASE_VERSION?=$(shell date +%Y%m%d%H%M)-v$(PROJECT_VERSION)
RELEASE_VERSION?=v$(PROJECT_VERSION)
SCHEDULER_RELEASE_IMAGE:=polaris-scheduler:$(RELEASE_VERSION)
CLUSTER_AGENT_RELEASE_IMAGE:=polaris-cluster-agent:$(RELEASE_VERSION)

# VERSION is the scheduler's version
#
# The RELEASE_VERSION variable can have one of two formats:
# v20201009-v0.18.800-46-g939c1c0 - automated build for a commit(not a tag) and also a local build
# v20200521-v0.18.800             - automated build for a tag
# VERSION=$(shell echo $(RELEASE_VERSION) | awk -F - '{print $$2}')
VERSION=v$(PROJECT_VERSION)

DEBUG_CONFIG=bin/config/kubernetes/scheduler.conf
DEBUG_CONFIG_PATH=$(shell echo "$(shell pwd)/$(DEBUG_CONFIG)" | sed -e "s|\/|\\\/|g" -)

.PHONY: all
all: build

.PHONY: build
build: build-scheduler build-cluster-agent

.PHONY: release-images
release-images: scheduler-release-image cluster-agent-release-image

.PHONY: clean
clean:
	rm -rf ./bin
	rm -rf ./vendor


###########################################
# polaris-scheduler
###########################################

.PHONY: build-scheduler
build-scheduler:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION) -w' -o bin/polaris-scheduler scheduler/main.go

.PHONY: build-scheduler-debug
build-scheduler-debug: scheduler-debug-config
	$(COMMONENVVAR) $(BUILDENVVAR) go build -gcflags="all=-N -l" -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION)' -o bin/polaris-scheduler scheduler/main.go

.PHONY: scheduler-debug-config
scheduler-debug-config:
	mkdir -p ./bin
	cp scheduler/manifests/polaris-scheduler/default-polaris-scheduler-config.yaml bin/default-polaris-scheduler-config.yaml

.PHONY: scheduler-release-image
scheduler-release-image:
	docker build -f ./scheduler/Dockerfile --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" -t $(RELEASE_REGISTRY)/$(SCHEDULER_RELEASE_IMAGE) .



###########################################
# polaris-cluster-agent
###########################################

.PHONY: build-cluster-agent
build-cluster-agent:
	$(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION) -w' -o bin/polaris-cluster-agent cluster-agent/main.go

.PHONY: build-cluster-agent-debug
build-cluster-agent-debug: cluster-agent-debug-config
	$(COMMONENVVAR) $(BUILDENVVAR) go build -gcflags="all=-N -l" -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION)' -o bin/polaris-cluster-agent cluster-agent/main.go

.PHONY: cluster-agent-debug-config
cluster-agent-debug-config:
	mkdir -p ./bin
	cp cluster-agent/manifests/default-polaris-cluster-agent-config.yaml bin/default-polaris-cluster-agent-config.yaml

.PHONY: cluster-agent-release-image
cluster-agent-release-image:
	docker build -f ./cluster-agent/Dockerfile --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" -t $(RELEASE_REGISTRY)/$(CLUSTER_AGENT_RELEASE_IMAGE) .
