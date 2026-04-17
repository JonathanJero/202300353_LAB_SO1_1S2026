from locust import HttpUser, task, between
from datetime import datetime, timezone
import os
import random

COUNTRIES = ["USA", "RUS", "CHN", "ESP", "GTM"]
CARNET = os.getenv("CARNET", "202300353")


class WarReporter(HttpUser):
    wait_time = between(0.2, 1.0)

    @task
    def send_report(self):
        payload = {
            "country": random.choice(COUNTRIES),
            "warplanes_in_air": random.randint(0, 50),
            "warships_in_water": random.randint(0, 30),
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        self.client.post(f"/grpc-{CARNET}", json=payload, timeout=5)
