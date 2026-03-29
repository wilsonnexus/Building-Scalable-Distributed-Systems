from locust import HttpUser, task, between
import random

SAMPLES = [
    "job failed due to timeout while calling dependency service",
    "flaky integration test failed intermittently",
    "connection refused while reaching internal API",
    "dependency download failed from artifact store",
    "unknown failure during pipeline execution"
]

class ReliabilityUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task
    def analyze_failure(self):
        payload = {
            "pipeline_id": f"pipe-{random.randint(1, 100000)}",
            "log_text": random.choice(SAMPLES),
            "source": "jenkins",
            "timestamp": "2026-03-28T12:00:00Z"
        }
        self.client.post("/analyze", json=payload, name="/analyze")