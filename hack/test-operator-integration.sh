#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

./multi-operator-manager test apply-configuration --test-dir=./test-data/apply-configuration/ --output-dir=./test-output --preserve-policy=KeepAlways