# Manual Técnico - Proyecto 3: M.U.M.N.K8s
**Monitoreo de Unidades Militares en la Nube con Kubernetes**
**Curso:** Laboratorio de Sistemas Operativos 1
**Estudiante:** 202300353 - Eliud Salguero

---

## Índice
1. [Arquitectura General](#1-arquitectura-general)
2. [Descripción de los Componentes](#2-descripción-de-los-componentes)
    - [Locust](#locust-el-simulador)
    - [Rust API](#rust-api)
    - [Go Ingest (Consumer gRPC)](#go-ingest-consumer-grpc)
    - [Go Writer (RabbitMQ Publisher)](#go-writer-rabbitmq-publisher)
    - [RabbitMQ](#rabbitmq-broker)
    - [Go Consumer (Writer Valkey)](#go-consumer)
3. [Virtualización Anidada (KubeVirt)](#3-virtualización-anidada-kubevirt)
    - [Valkey y Grafana en VMs](#valkey-y-grafana-en-vms)
4. [Infraestructura y Nube (GCP & Zot)](#4-infraestructura-y-nube-gcp--zot)
5. [Pruebas de Carga (Locust) y Autoescalado (HPA)](#5-pruebas-de-carga-locust-y-autoescalado-hpa)
6. [Grafana: Dashboard y País Asignado](#6-grafana-dashboard-y-país-asignado)
7. [Guía de Ejecución y Automatización](#7-guía-de-ejecución-y-automatización)

---

## 1. Arquitectura General

El proyecto plantea una **arquitectura distribuida basada en microservicios**, con el objetivo de recibir, traducir, encolar, procesar y almacenar de forma asincrónica miles de reportes militares enviados desde distintos frentes de batalla internacionales simulados. 

Toda la infraestructura se encuentra alojada sobre un clúster de **Google Kubernetes Engine (GKE)** y una Máquina Virtual separada utilizando de forma estricta familia de instancias **N1** de Google Cloud Platform (GCP).

### Flujo Exacto de Datos:
1. `Locust` genera reportes JSON concurrentes.
2. Todo el tráfico entra por el **Gateway API** de Kubernetes (Ruta: `/grpc-202300353`).
3. El `rust-api` intercepta el tráfico y envía la petición JSON al `go-ingest`.
4. El `go-ingest` traduce la petición en el protocolo binario ultrarrápido **gRPC** y la envía a `go-writer`.
5. El `go-writer` recibe la petición gRPC, la serializa, y publica el mensaje en una cola del broker **RabbitMQ**.
6. El `go-consumer` trabaja en _background_ desencolando ordenadamente en RabbitMQ.
7. Para cada mensaje desencolado, `go-consumer` lo registra permanentemente en la base de datos veloz **Valkey**.
8. **Grafana** consulta constantemente a Valkey para renderizar analíticas.

*(Nota: Valkey y Grafana no corren de forma nativa en Kubernetes, sino adentro de una virtualización anidada utilizando **KubeVirt** para emular hardware aislado).*

---

## 2. Descripción de los Componentes

### Locust (El simulador)
Aplicación Python encargada del bombardeo masivo de peticiones HTTP POST (Pruebas de Esfuerzo). Envía reportes con la siguiente estructura de datos (basado en un **OCI Artifact config.json** consumido desde Zot):
```json
{
    "country": "ESP",
    "warplanes_in_air": 42,
    "warships_in_water": 14,
    "timestamp": "2026-03-12T20:15:30Z"
}
```

### Rust API
Único servicio expuesto al internet (mediante Gateway API v1). Construido en `Rust` por su alto rendimiento (high throughput) y mínimo _footprint_ de memoria RAM. Su propósito es funcionar como el primer escudo ante la avalancha de eventos. Trabaja con límite de recursos asumiendo CPU objetivo de 30% para su _Horizontal Pod Autoscaler_ (1 a 3 réplicas). 

### Go Ingest (Consumer gRPC)
Servicio intermedio que toma el payload desde Rust y lo convierte al estándar **gRPC**. Implementamos Protocol Buffers para asegurar la velocidad en la transmisión interna de datos.
```protobuf
message WarReportRequest {
    Countries country = 1;
    int32 warplanes_in_air = 2;
    int32 warships_in_water = 3;
    string timestamp = 4;
}
```

### Go Writer (RabbitMQ Publisher)
Una vez que este servicio recibe el llamado gRPC, empaqueta el binario y establece una conexión asíncrona hacia AMQP (RabbitMQ). Este microservicio aisla a los clientes externos del peso que significa grabar datos en la base de datos primaria.

### RabbitMQ (Broker)
Sistema de colas implementado vía estado (Stateful) en Kubernetes. Nos garantiza **Zero Data Loss** en caso de que, durante un pico de Locust, Valkey se quede sin memoria o sea reiniciado. El Writer envía, el Consumer recoge.

### Go Consumer
Desplegado como un Worker perpetuo. Mantiene conexión en vivo a RabbitMQ y extrae mensaje por mensaje para ejecutar inserciones HTTP/Redis directas hacia la IP de la VM donde reside **Valkey**. 

---

## 3. Virtualización Anidada (KubeVirt)

Debido al requerimiento específico del enunciado, la capa de persistencia y visualización corren dentro de Máquinas Virtuales reales administradas por Kubernetes utilizando el operador **KubeVirt / CDI**.

### Valkey y Grafana en VMs
* **Valkey-VM**: Alojada bajo el CRD de `VirtualMachine` y un disco de `20Gi`, utilizando `Ubuntu:22.04` portado desde `containerdisks`. Mediante **CloudInit**, la VM instala en arranque Docker Engine e inicia un contenedor embebido de `valkey/valkey:8` puerto 6379 ligado al adaptador "masquerade" (`/k8s/kubevirt/valkey-vm.yaml`).
* **Grafana-VM**: Estructura similar a Valkey, pero su Docker Engine interno expone el puerto 3000 de `grafana/grafana:11.2.0` (`/k8s/kubevirt/grafana-vm.yaml`).

Ambas poseen la directriz `runStrategy: Always` garantizando que GKE las mantenga levantadas siempre.

---

## 4. Infraestructura y Nube (GCP & Zot)

* **GCP N1 Nodes**: El clúster de Kubernetes (`mumnk8s-gke`) está forzado estrictamente en su configuración de Terraform al uso de instancias `n1-standard-2` en el node-pool, garantizando el cumplimiento de la rúbrica sobre la arquitectura N1.
* **Zot Registry**: Implementado en GCP (fuera del clúster GKE) sobre una VM independiente. Funciona como un Docker Hub privado (OCI). Múltiples `make scripts` construyen los binarios de Rust y Go, generan la imagen Docker, y realizan el `docker push` a este Registry. Kubernetes usa estas URLs por defecto.
* **OCI Artifact**: Para demostrar la capacidad del registry en OCI, el proyecto almacena parámetros de configuración de combate de Locust (`config.json`) en Zot mediante la herramienta `oras`. En la inicialización, Locust hace un `pull` de este archivo y extrae los rangos aleatorios dinámicos demostrando el uso de artefactos no-Docker alojados en Zot.

---

## 5. Pruebas de Carga (Locust) y Autoescalado (HPA)

*(Espacio para evidencia estudiantil en defensa presencial)*
Durante la calificación presencial, se inyectarán cargas al `/grpc-202300353` corriendo Locust localmente con la directiva:
`make run-locust`

Esto disparará el evento del autómata **HPA (Horizontal Pod Autoscaler)**. Puesto que los deployments tienen estipulado:
```yaml
metrics:
- type: Resource
  resource:
    name: cpu
    target:
      type: Utilization
      averageUtilization: 30
```
Una vez Rust o Go Ingest sobrepasen el 30% de procesamiento en 1 réplica, Kubernetes aprovisionará dinámicamente hasta 3 réplicas para disipar la carga y evitar el HTTP 500 (Latencia / Drop).

**Pruebas de 1 vs 2 Réplicas en Go Writers:**
*(Anotar aquí en la presentación verbal: Al usar 1 réplica, la varianza/jitter de latencia frente a RabbitMQ puede aumentar y encolar más lento. Con 2 réplicas, balanceamos los TCP sockets contra RabbitMQ, procesando el doble de mensajes en el mismo segundo reduciendo lag en la gráfica temporal).*

---

## 6. Grafana: Dashboard y País Asignado

Dado el carnet **202300353**, el país asignado para seguimiento de serie temporal individual es **RUS (Rusia)**.
El Dashboard Final está configurado en Grafana con los siguientes paneles operando sobre **Valkey** (usando conector estándar Redis):
1. **Aviones en el Aire (General):** Valor máximo.
2. **Aviones en el Aire (General):** Valor mínimo.
3. **Barcos en Mar (General):** Valor máximo.
4. **Barcos en Mar (General):** Valor mínimo.
5. **Top Países:** Ranking tabular por mayor cantidad de aviones/barcos.
6. **Moda (Estadística):** El número de aviones y barcos que más veces se registró en la carga.
7. **Rusia (RUS):** Gráfica detallada _Time Series_ con la evolución histórica de aviones y barcos en el tiempo, acompañada de un _Stat_ que indique la Cantidad Total de Reportes Recibidos exclusivamente de este país.

*(Reemplazar acá con las capturas de pantalla de Grafana post-evaluación)*

---

## 7. Guía de Ejecución y Automatización

Para evitar gastos o consumo desmedido durante los laboratorios de prueba, este proyecto está altamente automatizado de inicio a fin utilizando contenedores Makefile. Todos los despliegues de la nube se levantan/apagan a discreción del usuario.

**Generar toda la infraestructura (Terraform & GKE):**
```bash
make infra-up
```
**Guardar saldo apagando temporalmente todos los servidores (Reducción a 0 nodos):**
```bash
make infra-stop-nodes
```
**Desplegar Kubernetes, KubeVirt, y Subir Imágenes a Zot:**
```bash
make build-push-images
make deploy-k8s
make deploy-kubevirt
make oci-artifact
```
**Levantar Locust y la prueba de carga:**
```bash
make run-locust
```
**Destruir completamente y apagar todo en Google Cloud:**
```bash
make infra-down
```
