#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}" )" && pwd)

BINARY=$1
GIT_REVISION="${GIT_REVISION:-$(git rev-parse --verify HEAD)}"
PROJECT_PKG="${PROJECT_PKG:-"github.com/nadundesilva/k8s-node-perf-evaluator"}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "${SCRIPT_DIR}/out/${BINARY}" "${SCRIPT_DIR}/cmd/${BINARY}"
