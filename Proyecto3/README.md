# Proyecto 3 - Infraestructura base en GCP

Este paquete te deja una base completa para levantar el proyecto del enunciado:

- Terraform para GKE + red + VM de Zot externa.
- Manifiestos Kubernetes para RabbitMQ, servicios base, HPA y Gateway API.
- Instalacion de KubeVirt + VMs de Valkey y Grafana.
- Locust para generar carga.
- Plantilla de manual tecnico.

## 1) Que tienes que hacer tu en Google Cloud Console

1. Crear un proyecto nuevo en GCP.
2. Activar Billing para ese proyecto.
3. En IAM, asignarte rol de Owner (o permisos equivalentes de Compute + GKE + IAM + Service Usage).
4. Instalar herramientas en tu maquina:
   - gcloud CLI
   - kubectl
   - terraform
   - docker
5. Hacer login:

```bash
gcloud auth login
```

## 2) Bootstrap inicial de GCP

Desde este directorio:

```bash
chmod +x scripts/*.sh k8s/kubevirt/install.sh
./scripts/bootstrap_gcp.sh TU_PROJECT_ID us-central1
```

## 3) Crear infraestructura con Terraform

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars
```

Edita `terraform.tfvars` y coloca como minimo:

- `project_id`
- `region`
- `zone`
- opcional recomendado: `zot_domain` + `zot_email` para TLS.

Aplica:

```bash
terraform init
terraform plan
terraform apply -auto-approve
```

Conecta kubectl a tu cluster:

```bash
gcloud container clusters get-credentials $(terraform output -raw cluster_name) \
  --region $(terraform output -raw cluster_region) \
  --project TU_PROJECT_ID
```

Activa Gateway API en GKE:

```bash
gcloud container clusters update $(terraform output -raw cluster_name) \
  --region $(terraform output -raw cluster_region) \
  --gateway-api=standard
```

## 4) Preparar Zot y publicar imagenes

Si usas dominio TLS, tu registry sera `zot.tudominio.com`.
Si no, saldra por IP:5000 (solo para pruebas iniciales).

Exporta registry:

```bash
export ZOT_REGISTRY="zot.tudominio.com"
```

Cuando tengas los Dockerfile de tus servicios:

```bash
./scripts/push_images.sh
```

## 5) Desplegar KubeVirt + VMs (Valkey y Grafana)

```bash
./scripts/deploy_kubevirt.sh
kubectl -n mumnk8s get vm,vmi,svc
```

## 6) Desplegar flujo principal del proyecto

```bash
export CARNET="202300353"
export ZOT_REGISTRY="zot.tudominio.com"
./scripts/deploy_k8s.sh
```

Verifica:

```bash
kubectl -n mumnk8s get pods,svc,hpa,gateway,httproute
```

## 7) Ejecutar Locust

Primero obten IP del Gateway (o LB):

```bash
kubectl -n mumnk8s get gateway
```

Luego:

```bash
cd locust
pip install locust
CARNET=202300353 locust -f locustfile.py --host http://IP_O_HOST_GATEWAY
```

## 8) Checklist rapido para que no te penalicen

- GKE activo con componentes funcionales.
- Locust enviando trafico real.
- RabbitMQ como broker principal.
- Valkey dentro de VM KubeVirt (independiente).
- Grafana dentro de VM KubeVirt (independiente).
- Gateway API con `/grpc-#carnet` (y `/dapr-#carnet` si haces extra).
- HPA en Rust de 1 a 3 replicas con CPU objetivo 30%.
- Imagenes publicadas y consumidas desde Zot.
- OCI Artifact documentado.
- Manual tecnico en Markdown y evidencia de pruebas.

## 9) Limitaciones de esta base

- Los manifiestos apuntan a imagenes `latest`; debes implementar y construir tus apps.
- El flujo Dapr es opcional y no se incluye aun.
- El dashboard final se construye en Grafana cuando ya tengas datos en Valkey.

## 10) Proximo paso sugerido

Implementar primero la API Rust y el Go ingest/writer para dejar funcional el flujo:

Locust -> Gateway -> Rust -> Go -> RabbitMQ

Luego integras consumer + Valkey + Grafana.
