from locust import HttpUser, task, between
import random
import uuid

SAMPLES = [
    "job failed due to timeout while calling dependency service",
    "flaky integration test failed intermittently",
    "connection refused while reaching internal API",
    "dependency download failed from artifact store",
    "unknown failure during pipeline execution"
]

class ReliabilityUser(HttpUser):
    wait_time = between(0.05, 0.2)

    @task
    def submit_event(self):
        payload = {
            "event_id": str(uuid.uuid4()),
            "pipeline_id": f"pipe-{random.randint(1, 100000)}",
            "log_text": random.choice(SAMPLES),
            "source": "jenkins",
            "timestamp": "2026-03-28T12:00:00Z"
        }
        self.client.post("/events", json=payload, name="/events")