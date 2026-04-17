#!/usr/bin/env bash
set -euo pipefail

KV_VERSION="${KV_VERSION:-v1.3.0}"
CDI_VERSION="${CDI_VERSION:-v1.60.0}"

echo "Instalando KubeVirt ${KV_VERSION}..."
kubectl apply -f "https://github.com/kubevirt/kubevirt/releases/download/${KV_VERSION}/kubevirt-operator.yaml"
kubectl apply -f "https://github.com/kubevirt/kubevirt/releases/download/${KV_VERSION}/kubevirt-cr.yaml"
kubectl -n kubevirt wait kv kubevirt --for=condition=Available --timeout=10m

echo "Instalando CDI ${CDI_VERSION}..."
kubectl apply -f "https://github.com/kubevirt/containerized-data-importer/releases/download/${CDI_VERSION}/cdi-operator.yaml"
kubectl apply -f "https://github.com/kubevirt/containerized-data-importer/releases/download/${CDI_VERSION}/cdi-cr.yaml"
kubectl -n cdi wait cdi cdi --for=condition=Available --timeout=10m

echo "KubeVirt y CDI instalados."
