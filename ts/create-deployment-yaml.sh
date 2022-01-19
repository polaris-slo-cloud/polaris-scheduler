#!/bin/bash
set -x
set -o errexit

OUTPUT="../deployment/4-slos.yaml"

# Delete old config.
rm -rf "$OUTPUT"


# YAML files to be combined into the output file.
INPUT_YAML_FILES=(
    # common-mappings
    "crds/customstreamsightslomappings.slo.k8s.rainbow-h2020.eu.yaml"
    "crds/migrationelasticitystrategies.elasticity.k8s.rainbow-h2020.eu.yaml"
    # "crds/imagethroughputslomappings.slo.k8s.rainbow-h2020.eu.yaml"

    # custom-stream-sight-slo-controller
    "apps/custom-stream-sight-slo-controller/manifests/kubernetes/1-rbac.yaml"
    "apps/custom-stream-sight-slo-controller/manifests/kubernetes/2-slo-controller.yaml"

    # migration-estrat-controller
    "apps/migration-estrat-controller/manifests/kubernetes/1-rbac.yaml"
    "apps/migration-estrat-controller/manifests/kubernetes/2-elasticity-strategy-controller.yaml"

    # image-throughput-slo-controller
    # "apps/image-throughput-slo-controller/manifests/kubernetes/1-rbac.yaml"
    # "apps/image-throughput-slo-controller/manifests/kubernetes/2-slo-controller.yaml"
)

for inputPath in ${INPUT_YAML_FILES[@]}; do
    cat "$inputPath" >> "$OUTPUT"
    echo -e "\n---\n" >> "$OUTPUT"
done

echo "Successfully wrote deployment config to $OUTPUT"
