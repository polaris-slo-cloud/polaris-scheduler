# Copyright 2020 The Kubernetes Authors.
# Modifications copyright 2020 Rainbow Project.
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

PROJECT_VERSION=0.0.14

LOCAL_REGISTRY=localhost:5000/rainbow
LOCAL_IMAGE=rainbow-scheduler:latest

# TODO: This needs to be adjusted for the RAINBOW registry.
# RELEASE_REGISTRY is the container registry to push
# into. The default is to push to the staging
# registry, not production(k8s.gcr.io).
RELEASE_REGISTRY?=rainbowh2020
RELEASE_VERSION?=$(shell date +%Y%m%d%H%M)-v$(PROJECT_VERSION)
# RELEASE_VERSION?=v$(shell date +%Y%m%d)-$(shell git describe --tags --match "v*")
RELEASE_IMAGE:=rainbow-scheduler:$(RELEASE_VERSION)

# VERSION is the scheduler's version
#
# The RELEASE_VERSION variable can have one of two formats:
# v20201009-v0.18.800-46-g939c1c0 - automated build for a commit(not a tag) and also a local build
# v20200521-v0.18.800             - automated build for a tag
VERSION=$(shell echo $(RELEASE_VERSION) | awk -F - '{print $$2}')

DEBUG_CONFIG=bin/config/kubernetes/scheduler.conf
DEBUG_CONFIG_PATH=$(shell echo "$(shell pwd)/$(DEBUG_CONFIG)" | sed -e "s|\/|\\\/|g" -)

.PHONY: all
all: build

.PHONY: build
build: build-scheduler

.PHONY: build-scheduler
build-scheduler: autogen
	$(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION) -w' -o bin/rainbow-scheduler cmd/scheduler/main.go

.PHONY: build-scheduler-debug
build-scheduler-debug: autogen debug-config
	$(COMMONENVVAR) $(BUILDENVVAR) go build -gcflags="all=-N -l" -ldflags '-X k8s.io/component-base/version.gitVersion=$(VERSION)' -o bin/rainbow-scheduler cmd/scheduler/main.go

.PHONY: local-image
local-image:
	docker build -f ./build/scheduler/Dockerfile --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" -t $(LOCAL_REGISTRY)/$(LOCAL_IMAGE) ..

.PHONY: release-image
release-image:
	docker build -f ./build/scheduler/Dockerfile --build-arg RELEASE_VERSION="$(RELEASE_VERSION)" -t $(RELEASE_REGISTRY)/$(RELEASE_IMAGE) ..

# .PHONY: push-release-image
# push-release-image: release-image
# 	gcloud auth configure-docker
# 	docker push $(RELEASE_REGISTRY)/$(RELEASE_IMAGE)

.PHONY: update-vendor
update-vendor:
	hack/update-vendor.sh

.PHONY: unit-test
unit-test: autogen
	hack/unit-test.sh

.PHONY: install-etcd
install-etcd:
	hack/install-etcd.sh

.PHONY: autogen
autogen: update-vendor
	hack/update-generated-openapi.sh

.PHONY: integration-test
integration-test: install-etcd autogen
	hack/integration-test.sh

.PHONY: verify-gofmt
verify-gofmt:
	hack/verify-gofmt.sh

.PHONY: debug-config
debug-config:
	mkdir -p ./bin
	sed -e "s/kubeconfig:.*/kubeconfig: '$(DEBUG_CONFIG_PATH)'/" manifests/rainbow-scheduler/default-rainbow-scheduler-config.yaml > bin/default-rainbow-scheduler-config.yaml

.PHONY: clean
clean:
	rm -rf ./bin
	rm -rf ./vendor
