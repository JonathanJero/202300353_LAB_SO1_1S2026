#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-202300353}"

if [[ -z "${ZOT_REGISTRY:-}" ]]; then
  echo "Define ZOT_REGISTRY antes de ejecutar."
  echo "Ejemplo: ZOT_REGISTRY=zot.midominio.com ./scripts/deploy_k8s.sh"
  exit 1
fi

RABBITMQ_USER="${RABBITMQ_USER:-admin}"
RABBITMQ_PASSWORD="${RABBITMQ_PASSWORD:-}"

kubectl apply -f k8s/base/namespace.yaml

if ! kubectl -n mumnk8s get secret rabbitmq-auth >/dev/null 2>&1; then
  if [[ -z "$RABBITMQ_PASSWORD" ]]; then
    RABBITMQ_PASSWORD="$(openssl rand -hex 16)"
    echo "Se genero password aleatorio para RabbitMQ (no se guarda en git)."
  fi

  RABBITMQ_URL="amqp://${RABBITMQ_USER}:${RABBITMQ_PASSWORD}@rabbitmq.mumnk8s.svc.cluster.local:5672/"

  kubectl -n mumnk8s create secret generic rabbitmq-auth \
    --from-literal=username="$RABBITMQ_USER" \
    --from-literal=password="$RABBITMQ_PASSWORD" \
    --from-literal=rabbitmq_url="$RABBITMQ_URL"
else
  echo "Secret rabbitmq-auth ya existe, se reutiliza para evitar rotaciones accidentales."
fi

kubectl apply -f k8s/base/rabbitmq.yaml

envsubst < k8s/base/apps.tmpl.yaml | kubectl apply -f -
kubectl apply -f k8s/base/hpa.yaml
envsubst < k8s/base/gateway.tmpl.yaml | kubectl apply -f -

echo "Despliegue base aplicado."
echo "CARNET utilizado: ${CARNET}"
echo "Verifica Gateway con: kubectl -n mumnk8s get gateway,httproute,svc,pods"
