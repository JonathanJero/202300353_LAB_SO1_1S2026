#!/usr/bin/env bash
set -euo pipefail

./k8s/kubevirt/install.sh
kubectl apply -f k8s/kubevirt/valkey-vm.yaml
kubectl apply -f k8s/kubevirt/grafana-vm.yaml
kubectl apply -f k8s/kubevirt/vm-services.yaml

echo "KubeVirt + VMs desplegadas."
echo "Revisa estado con: kubectl -n mumnk8s get vmi,vm,svc"
