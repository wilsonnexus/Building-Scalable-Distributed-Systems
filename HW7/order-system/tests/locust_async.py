from locust import HttpUser, task, between
import random

def make_order():
    return {
        "customer_id": random.randint(1000, 9999),
        "items": [
            {"item_id": 1, "name": "shirt", "quantity": 1},
            {"item_id": 2, "name": "hat", "quantity": 2}
        ]
    }

class AsyncOrderUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task
    def create_async_order(self):
        self.client.post("/orders/async", json=make_order(), name="/orders/async")