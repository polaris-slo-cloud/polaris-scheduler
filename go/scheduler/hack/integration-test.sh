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

checkEtcdOnPath() {
  # If it's in a prow CI env, add etcd to path.
  [[ ${CI:-} == "true" ]] && export PATH="$(pwd)/etcd:${PATH}"
  kube::log::status "Checking etcd is on PATH"
  command -v etcd >/dev/null && return
  kube::log::status "Cannot find etcd, cannot run integration tests."
  kube::log::status "Please see https://git.k8s.io/community/contributors/devel/sig-testing/integration-tests.md#install-etcd-dependency for instructions."
  # kube::log::usage "You can use 'hack/install-etcd.sh' to install a copy in third_party/."
  return 1
}

CLEANUP_REQUIRED=
cleanup() {
  [[ -z "${CLEANUP_REQUIRED}" ]] && return
  kube::log::status "Cleaning up etcd"
  kube::etcd::cleanup
  CLEANUP_REQUIRED=
  kube::log::status "Integration test cleanup complete"
}

runTests() {
  kube::log::status "Starting etcd instance"
  CLEANUP_REQUIRED=1
  kube::etcd::start
  kube::log::status "Running integration test cases"

  # TODO: make args customizable.
  go test -mod=vendor polaris-slo-cloud.github.io/polaris-scheduler/v1/scheduler/test/integration/...

  cleanup
}

checkEtcdOnPath

# Run cleanup to stop etcd on interrupt or other kill signal.
trap cleanup EXIT

runTests
