# Setup local (macOS o Linux)

Este proyecto se puede desarrollar desde macOS o Linux. No necesitas migrarte obligatoriamente a Linux.

## Opcion A: macOS (recomendado si ya trabajas comodo ahi)

1. Instalar Homebrew (si no lo tienes):

   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

2. Instalar herramientas base:

   brew install --cask google-cloud-sdk
   brew install kubectl
   brew tap hashicorp/tap
   brew install hashicorp/tap/terraform
   brew install --cask docker

   Si aparece error de formula no disponible para terraform, ejecuta:

   brew update
   brew install hashicorp/tap/terraform

3. Iniciar Docker Desktop una vez para habilitar el daemon.

4. Verificar:

   gcloud --version
   kubectl version --client
   terraform version
   docker --version

## Opcion B: Linux (Ubuntu/Debian VM)

1. Actualizar e instalar utilidades:

   sudo apt update
   sudo apt install -y curl ca-certificates gnupg lsb-release apt-transport-https software-properties-common

2. Instalar Google Cloud CLI:

   curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/cloud.google.gpg
   echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee /etc/apt/sources.list.d/google-cloud-sdk.list
   sudo apt update && sudo apt install -y google-cloud-cli

3. Instalar kubectl y Terraform:

   sudo apt install -y kubectl
   curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
   echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
   sudo apt update && sudo apt install -y terraform

4. Instalar Docker:

   curl -fsSL https://get.docker.com | sh
   sudo usermod -aG docker $USER
   newgrp docker

5. Verificar:

   gcloud --version
   kubectl version --client
   terraform version
   docker --version

## Seguridad recomendada

- Usa gcloud auth login y gcloud auth application-default login.
- No subas llaves JSON de service account al repositorio.
- No guardes secretos en YAML versionados.
- Usa variables de entorno o Kubernetes Secret creado en runtime.

## Tu carnet en este proyecto

- Carnet configurado por defecto: 202300353
- Ruta principal: /grpc-202300353
- Ruta opcional: /dapr-202300353
