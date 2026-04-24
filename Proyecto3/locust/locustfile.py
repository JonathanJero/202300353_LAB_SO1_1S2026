from locust import HttpUser, task, between
from datetime import datetime, timezone
import os
import random

import json

COUNTRIES = ["USA", "RUS", "CHN", "ESP", "GTM"]
CARNET = os.getenv("CARNET", "202300353")

# Cargar configuracion descargada como OCI Artifact si existe
max_planes = 50
max_ships = 30
if os.path.exists("config.json"):
    with open("config.json", "r") as f:
        conf = json.load(f)
        max_planes = conf.get("warplanes_max", 50)
        max_ships = conf.get("warships_max", 30)

class WarReporter(HttpUser):
    wait_time = between(0.2, 1.0)

    @task
    def send_report(self):
        payload = {
            "country": random.choice(COUNTRIES),
            "warplanes_in_air": random.randint(0, max_planes),
            "warships_in_water": random.randint(0, max_ships),
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        self.client.post(f"/grpc-{CARNET}", json=payload, timeout=5)
