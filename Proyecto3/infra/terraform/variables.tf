variable "project_id" {
  description = "ID del proyecto de GCP"
  type        = string
}

variable "region" {
  description = "Region de GCP"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "Zona principal"
  type        = string
  default     = "us-central1-a"
}

variable "name_prefix" {
  description = "Prefijo para nombrar recursos"
  type        = string
  default     = "mumnk8s"
}

variable "gke_name" {
  description = "Nombre del cluster de GKE"
  type        = string
  default     = "mumnk8s-gke"
}

variable "network_cidr" {
  description = "CIDR de la subred"
  type        = string
  default     = "10.10.0.0/16"
}

variable "node_count" {
  description = "Cantidad inicial de nodos"
  type        = number
  default     = 2
}

variable "node_machine_type" {
  description = "Tipo de maquina para nodos GKE (N1 requerido por enunciado)"
  type        = string
  default     = "n1-standard-2"
}

variable "zot_machine_type" {
  description = "Tipo de maquina para VM de Zot"
  type        = string
  default     = "n1-standard-1"
}

variable "zot_disk_size_gb" {
  description = "Disco de VM Zot"
  type        = number
  default     = 50
}

variable "admin_cidr" {
  description = "CIDR desde donde permites SSH a la VM"
  type        = string
  default     = "0.0.0.0/0"
}

variable "zot_domain" {
  description = "Dominio publico para Zot (opcional, recomendado para TLS)"
  type        = string
  default     = ""
}

variable "zot_email" {
  description = "Email para certificados TLS automaticos (si usas dominio)"
  type        = string
  default     = ""
}
