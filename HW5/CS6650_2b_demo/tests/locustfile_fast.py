from locust import FastHttpUser, task, between

import random

class AlbumUser(FastHttpUser):
    wait_time = between(0.01, 0.05)

    @task(9)
    def list_albums(self):
        self.client.get("/albums")

    @task(1)
    def get_album(self):
        album_id = random.choice(["1", "2", "3"])
        self.client.get(f"/albums/{album_id}")
