# Kubernetes Cluster Nodes' Performance Evaluator

[![Build](https://github.com/nadundesilva/k8s-node-perf-evaluator/actions/workflows/build-perf-evaluator.yaml/badge.svg)](https://github.com/nadundesilva/k8s-node-perf-evaluator/actions/workflows/build-perf-evaluator.yaml)
[![Code Scan](https://github.com/nadundesilva/k8s-node-perf-evaluator/actions/workflows/code-scan.yaml/badge.svg)](https://github.com/nadundesilva/k8s-node-perf-evaluator/actions/workflows/code-scan.yaml)

This repository contains a set of tools for testing the performance of all the nodes in a kubernetes cluster. When using nodes provided by cloud providers, there are cases where there some nodes in the cluster which are performing badly. This tool provides a way to test the performance of all the nodes in the cluster.

## Supported Test Types

* Ping test
* CPU intensive load test

## How to Use

### Updating Configurations

1. Clone this repository and navigate to the root of the directory.
2. Update `config.yaml` with the proper configurations about the cluster (The exact configurations will change based on the method of running the test as well as the cluster).

### How to run Test

#### Run using Docker Image

Run the following command to execute tests
```bash
docker run --name=k8s-node-performance-evaluator \
    --rm \
    --volume=${PWD}/config.yaml:/app/config.yaml:ro \
    --volume=${HOME}/.kube/config:/.kube/config:ro \
    ghcr.io/nadundesilva/tools/k8s-node-perf-evaluator/test-runner:latest
```

#### Build and Run from Source

Run the following command to execute tests
```bash
go run cmd/test-runner/main.go
```
