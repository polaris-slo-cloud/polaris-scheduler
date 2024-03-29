#!/usr/bin/env bash

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

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${SCRIPT_ROOT}/hack/lib/init.sh"

# TODO: make args customizable.
go test -mod=vendor \
  polaris-slo-cloud.github.io/polaris-scheduler/v1/scheduler/cmd/... \
  polaris-slo-cloud.github.io/polaris-scheduler/v1/scheduler/pkg/...
