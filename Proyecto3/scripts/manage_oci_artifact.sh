#!/usr/bin/env bash
set -euo pipefail

ZOT=$1
if [[ -z "$ZOT" ]]; then
  echo "Uso: $0 <ZOT_REGISTRY>"
  exit 1
fi

echo ">> Preparando ORAS (OCI Registry As Storage CLI)..."
# Descargar ORAS temporalmente si no existe en el sistema
if ! command -v oras &> /dev/null; then
    echo "Descargando oras cli..."
    curl -sLO "https://github.com/oras-project/oras/releases/download/v1.1.0/oras_1.1.0_windows_amd64.zip" || curl -sLO "https://github.com/oras-project/oras/releases/download/v1.1.0/oras_1.1.0_linux_amd64.tar.gz"
    tar -zxf oras_1.1.0_*.tar.gz -C /usr/local/bin/ 2>/dev/null || unzip -q oras_1.1.0_windows_amd64.zip -d oras-win || true
    export PATH=$PATH:$(pwd)/oras-win
fi

echo ">> Empujando locust/config.json como OCI Artifact a $ZOT..."
cd locust

# Deshabilitar TLS verification si ZOT no tiene certificados válidos 
export ORAS_PING=false

# Subir a ZOT
oras push $ZOT/mumnk8s/config:1.0 --plain-http \
  --config /dev/null:application/vnd.mumnk8s.config.v1+json \
  config.json:application/json

echo ">> !Exito! Archivo subido. Ahora probando la descarga del OCI Artifact..."
mkdir -p /tmp/oci-test
cd /tmp/oci-test
oras pull $ZOT/mumnk8s/config:1.0 --plain-http

echo ">> Contenido descargado desde ZOT:"
cat config.json
echo ""
echo ">> OCI Artifact publicado y descargado con exito. Muestra esta prueba a tu calificador."
