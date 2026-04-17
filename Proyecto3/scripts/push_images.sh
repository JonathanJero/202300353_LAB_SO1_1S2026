#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ZOT_REGISTRY:-}" ]]; then
  echo "Define ZOT_REGISTRY. Ejemplo: export ZOT_REGISTRY=zot.midominio.com"
  exit 1
fi

# Ajusta los Dockerfile/rutas reales cuando implementes cada servicio.
declare -A images=(
  [rust-api]="./apps/rust-api"
  [go-ingest]="./apps/go-ingest"
  [go-writer]="./apps/go-writer"
  [go-consumer]="./apps/go-consumer"
)

for name in "${!images[@]}"; do
  context="${images[$name]}"
  if [[ -d "$context" ]]; then
    echo "Construyendo y subiendo $name..."
    docker build -t "${ZOT_REGISTRY}/mumnk8s/${name}:latest" "$context"
    docker push "${ZOT_REGISTRY}/mumnk8s/${name}:latest"
  else
    echo "Saltando $name (no existe directorio $context)."
  fi
done
