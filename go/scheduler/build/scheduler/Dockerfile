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
FROM golang:1.16

# Important: The Docker context's root directory is not the `go/scheduler` directory,
# but the `go` directory instead, because the scheduler depends on the orchestrator module.

WORKDIR /go/src/k8s.rainbow-h2020.eu/rainbow
COPY . .

WORKDIR /go/src/k8s.rainbow-h2020.eu/rainbow/scheduler
ARG RELEASE_VERSION
RUN RELEASE_VERSION=${RELEASE_VERSION} make build-scheduler

FROM alpine:3.12

COPY ./scheduler/manifests/rainbow-scheduler/default-rainbow-scheduler-config.yaml /etc/rainbow-scheduler/default-rainbow-scheduler-config.yaml
COPY --from=0 /go/src/k8s.rainbow-h2020.eu/rainbow/scheduler/bin/rainbow-scheduler /bin/rainbow-scheduler

WORKDIR /bin
CMD ["rainbow-scheduler"]
