# Terraform de infraestructura base

Este directorio crea:

- VPC y subred dedicadas.
- Cluster de GKE (Standard) con nodos N1.
- VM externa para Zot (fuera del cluster).

## Requisitos locales

- gcloud CLI autenticado.
- Terraform >= 1.5.

## Uso rapido

1. Copia el archivo de ejemplo:

   cp terraform.tfvars.example terraform.tfvars

2. Edita `terraform.tfvars` con tu `project_id`.

3. Inicializa Terraform:

   terraform init

4. Previsualiza:

   terraform plan

5. Aplica:

   terraform apply -auto-approve

6. Obtiene credenciales de kubectl:

    gcloud container clusters get-credentials $(terraform output -raw cluster_name) \
       --location $(terraform output -raw cluster_location) \
     --project <tu-project-id>

## Notas de Zot

- Si defines `zot_domain` y `zot_email`, se levanta Caddy para TLS automatico en 443.
- Si no defines dominio, Zot queda en HTTP por puerto 5000 (util para pruebas iniciales, pero no recomendado para pulls de Kubernetes en produccion).
