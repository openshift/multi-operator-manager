#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPLACE_TEST_OUTPUT="${REPLACE_TEST_OUTPUT:-false}"

OUTPUT_DIR="${ARTIFACT_DIR:-.}"/test-output
rm -rf "${OUTPUT_DIR}"

if [ "$REPLACE_TEST_OUTPUT" == "true" ]
then
  ./multi-operator-manager test apply-configuration --test-dir=./test-data/apply-configuration/ --output-dir="${OUTPUT_DIR}" --replace-expected-output=true
else
  ./multi-operator-manager test apply-configuration --test-dir=./test-data/apply-configuration/ --output-dir="${OUTPUT_DIR}" --preserve-policy=KeepAlways
fi