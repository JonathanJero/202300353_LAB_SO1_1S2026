# Manual Tecnico - Proyecto 3 M.U.M.N.K8s

## 1. Arquitectura general

- Diagrama del flujo completo.
- Componentes obligatorios implementados.
- Justificacion de decisiones tecnicas.

## 2. Flujo de datos

- Locust -> Gateway API -> Rust -> Go ingest -> gRPC writer -> RabbitMQ -> Consumer -> Valkey -> Grafana.
- Formato del mensaje JSON.
- Contrato gRPC (proto).

## 3. Infraestructura en GCP

- Proyecto, region, zona y tipo de instancias N1.
- Evidencia de cluster GKE.
- Evidencia de VM de Zot fuera del cluster.

## 4. Gateway API

- GatewayClass utilizada.
- Rutas implementadas:
  - /grpc-#carnet
  - /dapr-#carnet (si aplica)

## 5. Servicios y comunicacion

- API REST en Rust.
- Servicios de Go:
  - Deployment 1: API REST + gRPC client.
  - Deployment 2/3: gRPC server + writer RabbitMQ.
- Broker RabbitMQ y colas.

## 6. KubeVirt

- Instalacion de KubeVirt y CDI.
- VM de Valkey independiente.
- VM de Grafana independiente.
- Evidencia de conectividad y persistencia.

## 7. HPA y recursos

- Configuracion de requests/limits.
- HPA de Rust (1-3 replicas, CPU objetivo 30%).
- Evidencias de escalamiento.

## 8. Zot y OCI Artifact

- Publicacion y consumo de imagenes Docker desde Zot.
- OCI Artifact usado (archivo, comando de push/pull y uso).

## 9. Pruebas de carga

- Configuracion de Locust.
- Escenarios y volumen de carga.
- Pruebas comparativas con 1 y 2 replicas (componentes requeridos).
- Resultados y analisis.

## 10. Dashboard en Grafana

- Visualizaciones obligatorias implementadas.
- Asignacion de pais segun ultimo digito de carnet.
- Capturas y conclusiones.

## 11. Conclusiones

- Lecciones aprendidas.
- Limites detectados.
- Mejoras propuestas.
