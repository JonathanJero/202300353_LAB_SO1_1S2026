#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-202300353}"

if [[ -z "${PROJECT_ID:-}" || -z "${REGION:-}" || -z "${ZOT_REGISTRY:-}" ]]; then
  echo "Define PROJECT_ID, REGION y ZOT_REGISTRY antes de ejecutar."
  echo "Ejemplo: PROJECT_ID=mi-proyecto REGION=us-central1 ZOT_REGISTRY=zot.midominio.com ./scripts/deploy_all.sh"
  exit 1
fi

pushd infra/terraform >/dev/null
terraform init
terraform apply -auto-approve
CLUSTER_NAME="$(terraform output -raw cluster_name)"
CLUSTER_REGION="$(terraform output -raw cluster_region)"
popd >/dev/null

gcloud container clusters get-credentials "$CLUSTER_NAME" --region "$CLUSTER_REGION" --project "$PROJECT_ID"
gcloud container clusters update "$CLUSTER_NAME" --region "$CLUSTER_REGION" --gateway-api=standard

./scripts/deploy_kubevirt.sh
./scripts/deploy_k8s.sh

echo "Infraestructura y manifests aplicados."
