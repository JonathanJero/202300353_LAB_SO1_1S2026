PROJECT_ID ?= mumnk8s-202300353
REGION ?= us-central1
ZONE ?= us-central1-f
CARNET ?= 202300353
ZOT_REGISTRY ?= localhost:5000

.PHONY: help infra-up infra-down infra-stop-nodes k8s-creds build-push-images deploy-k8s deploy-kubevirt run-locust oci-artifact

help:
	@echo "M.U.M.N.K8s - Comandos de Automatización del Proyecto 3"
	@echo "-------------------------------------------------------"
	@echo "make infra-up          - Levantar infraestructura en GCP (Terraform)"
	@echo "make infra-stop-nodes  - Reducir nodos a 0 para no gastar saldo"
	@echo "make k8s-creds         - Obtener credenciales de Kubernetes (kubectl)"
	@echo "make build-push-images - Construir imágenes Docker y subirlas a Zot"
	@echo "make oci-artifact      - Cumplir req OCI: Subir y descargar config.json a Zot"
	@echo "make deploy-k8s        - Desplegar manifiestos de K8s (Gateway, Deployments)"
	@echo "make deploy-kubevirt   - Instalar KubeVirt y crear VMs de Valkey y Grafana"
	@echo "make run-locust        - Iniciar Locust para generar carga"
	@echo "make infra-down        - Destruir toda la infraestructura en GCP"

infra-up:
	cd infra/terraform && terraform init && terraform apply -auto-approve

infra-down:
	cd infra/terraform && terraform destroy -auto-approve

infra-stop-nodes:
	gcloud container clusters resize mumnk8s-gke --project $(PROJECT_ID) --zone $(ZONE) --node-pool mumnk8s-node-pool --num-nodes 0 --quiet

k8s-creds:
	gcloud container clusters get-credentials mumnk8s-gke --location $(ZONE) --project $(PROJECT_ID)

build-push-images:
	ZOT_REGISTRY=$(ZOT_REGISTRY) ./scripts/push_images.sh

oci-artifact:
	chmod +x scripts/manage_oci_artifact.sh
	./scripts/manage_oci_artifact.sh $(ZOT_REGISTRY)

deploy-k8s:
	CARNET=$(CARNET) envsubst < k8s/base/gateway.tmpl.yaml | kubectl apply -f -
	kubectl apply -f k8s/base/namespace.yaml
	kubectl apply -f k8s/base/rabbitmq.yaml
	kubectl apply -f k8s/api/deployments.yaml

deploy-kubevirt:
	chmod +x k8s/kubevirt/install.sh
	./k8s/kubevirt/install.sh
	kubectl apply -f k8s/kubevirt/valkey-vm.yaml
	kubectl apply -f k8s/kubevirt/grafana-vm.yaml
	kubectl apply -f k8s/kubevirt/vm-services.yaml

run-locust:
	cd locust && pip install locust && locust -f locustfile.py
