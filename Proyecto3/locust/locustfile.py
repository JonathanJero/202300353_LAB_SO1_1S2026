from locust import HttpUser, task, between
from datetime import datetime, timezone
import os
import random

import json
import threading

COUNTRIES = ["usa", "rus", "chn", "esp", "gtm"]
CARNET = os.getenv("CARNET", "202300353")

# Estado global para mantener datos coherentes (Random Walk)
state_lock = threading.Lock()
country_state = {
    "usa": {"planes": 850, "ships": 1200},
    "rus": {"planes": 620, "ships": 950},
    "chn": {"planes": 410, "ships": 320},
    "esp": {"planes": 380, "ships": 780},
    "gtm": {"planes": 210, "ships": 450},
}

class WarReporter(HttpUser):
    wait_time = between(0.2, 1.0)

    @task
    def send_report(self):
        c = random.choice(COUNTRIES)
        
        with state_lock:
            # Caminata aleatoria suave (+- 5)
            country_state[c]["planes"] += random.randint(-5, 5)
            country_state[c]["ships"] += random.randint(-5, 5)
            
            # Evitar números negativos
            if country_state[c]["planes"] < 0: country_state[c]["planes"] = 0
            if country_state[c]["ships"] < 0: country_state[c]["ships"] = 0
            
            planes = country_state[c]["planes"]
            ships = country_state[c]["ships"]

        payload = {
            "country": c,
            "warplanes_in_air": planes,
            "warships_in_water": ships,
            "timestamp": datetime.now(timezone.utc).isoformat(),
        }
        self.client.post(f"/grpc-{CARNET}", json=payload, timeout=5)
