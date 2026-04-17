output "cluster_name" {
  value = google_container_cluster.main.name
}

output "cluster_region" {
  value = var.region
}

output "zot_public_ip" {
  value = google_compute_address.zot_ip.address
}

output "zot_registry_http" {
  value = "${google_compute_address.zot_ip.address}:5000"
}

output "zot_registry_https" {
  value = var.zot_domain != "" ? var.zot_domain : "(no domain configured)"
}
