from locust import FastHttpUser, task, between
import random
import time
-
class AlbumsUser(FastHttpUser):
    wait_time = between(0.1, 0.5)

    @task(3)
    def get_albums(self):
        self.client.get("/albums", name="GET /albums")

    @task(1)
    def post_album(self):
        # Unique ID so you don't hit "duplicate" validation
        unique_id = f"locust-{int(time.time()*1000)}-{random.randint(0, 999999)}"

        payload = {
            "id": unique_id,
            "title": "Load Test Album",
            "artist": "Locust",
            "price": 1.23
        }

        # catch_response lets us mark failures explicitly + see why
        with self.client.post("/albums", json=payload, name="POST /albums", catch_response=True) as resp:
            if resp.status_code >= 400:
                resp.failure(f"{resp.status_code} {resp.text}")
            else:
                resp.success()
