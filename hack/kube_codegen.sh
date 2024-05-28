#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# Set up the environment variables
CRD_GROUP_VERSION="kubelitedb:v1"
MODULE=github.com/robert-cronin/kubelitedb
OUTPUT_PKG=$MODULE/pkg/generated
APIS_PKG=$MODULE/pkg/apis
GROUP_VERSION=kubelitedb:v1

# Run the code generator using the vendored code-generator
bash vendor/k8s.io/code-generator/kube_codegen.sh all \
    $OUTPUT_PKG $APIS_PKG $GROUP_VERSION \
    --go-header-file "./boilerplate.go.txt"
