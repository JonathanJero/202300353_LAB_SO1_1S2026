resource "google_compute_address" "zot_ip" {
  name   = "${var.name_prefix}-zot-ip"
  region = var.region
}

resource "google_compute_instance" "zot" {
  name         = "${var.name_prefix}-zot-vm"
  machine_type = var.zot_machine_type
  zone         = var.zone
  tags         = ["${var.name_prefix}-zot"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = var.zot_disk_size_gb
      type  = "pd-standard"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id

    access_config {
      nat_ip = google_compute_address.zot_ip.address
    }
  }

  metadata_startup_script = templatefile("${path.module}/startup-zot.sh.tftpl", {
    zot_domain = var.zot_domain
    zot_email  = var.zot_email
  })

  depends_on = [google_project_service.required]
}
