#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

OUTPUT_DIR="${ARTIFACT_DIR:-.}"/test-output
rm -rf "${OUTPUT_DIR}"
./multi-operator-manager test apply-configuration --test-dir=./test-data/apply-configuration/ --output-dir="${OUTPUT_DIR}" --preserve-policy=KeepAlways