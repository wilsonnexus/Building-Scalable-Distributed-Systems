from locust import FastHttpUser, task, constant
import random

SEARCH_TERMS = [
    "electronics", "books", "home", "clothing", "sports", "toys",
    "alpha", "bravo", "nova", "zen", "product"
]

class ProductSearchUser(FastHttpUser):
    wait_time = constant(0)  # minimal wait time

    @task
    def search(self):
        q = random.choice(SEARCH_TERMS)
        self.client.get(f"/products/search?q={q}", name="/products/search")
