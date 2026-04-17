#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  echo "Uso: $0 <PROJECT_ID> <REGION>"
  exit 1
fi

PROJECT_ID="$1"
REGION="$2"

gcloud auth login
gcloud auth application-default login
gcloud config set project "$PROJECT_ID"
gcloud config set compute/region "$REGION"

gcloud services enable \
  compute.googleapis.com \
  container.googleapis.com \
  iam.googleapis.com \
  serviceusage.googleapis.com \
  cloudresourcemanager.googleapis.com

echo "Proyecto y APIs configurados en GCP."
echo "Se uso autenticacion de usuario (sin llaves JSON en el repo)."
