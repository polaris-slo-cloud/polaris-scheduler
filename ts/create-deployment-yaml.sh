#!/bin/bash
set -x
set -o errexit

# ToDo: Add possibility to specify version, because currently the version must be manually set in th econtrollers' YAML files.

OUTPUT="../deployment/4-slos.yaml"

# Delete old config.
rm -rf "$OUTPUT"


# Build the projects.
INPUT_YAML_FILES=(
    # image-throughput-slo-controller
    "apps/image-throughput-slo-controller/manifests/kubernetes/1-rbac.yaml"
    "apps/image-throughput-slo-controller/manifests/kubernetes/2-slo-controller.yaml"
)

for inputPath in ${INPUT_YAML_FILES[@]}; do
    cat "$inputPath" >> "$OUTPUT"
    echo -e "\n---\n" >> "$OUTPUT"
done

echo "Successfully wrote deployment config to $OUTPUT"
