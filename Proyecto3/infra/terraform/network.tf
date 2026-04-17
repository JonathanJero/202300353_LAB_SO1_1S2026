resource "google_compute_network" "main" {
  name                    = "${var.name_prefix}-vpc"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "main" {
  name          = "${var.name_prefix}-subnet"
  ip_cidr_range = var.network_cidr
  region        = var.region
  network       = google_compute_network.main.id
}

resource "google_compute_firewall" "zot_allow_ssh" {
  name    = "${var.name_prefix}-zot-allow-ssh"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = [var.admin_cidr]
  target_tags   = ["${var.name_prefix}-zot"]
}

resource "google_compute_firewall" "zot_allow_http" {
  name    = "${var.name_prefix}-zot-allow-http"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["5000", "80", "443"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["${var.name_prefix}-zot"]
}
