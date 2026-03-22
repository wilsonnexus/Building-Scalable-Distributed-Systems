import json
import random
import requests
import time
from datetime import datetime, timezone

BASE_URL = "http://cartapi-alb-1350600380.us-east-1.elb.amazonaws.com"
RESULTS_FILE = "mysql_test_results.json"

results = []
cart_ids = []

def now_iso():
    return datetime.now(timezone.utc).isoformat()

def record(operation, response_time_ms, success, status_code):
    results.append({
        "operation": operation,
        "response_time": round(response_time_ms, 2),
        "success": success,
        "status_code": status_code,
        "timestamp": now_iso()
    })

# 50 create cart
for i in range(50):
    payload = {
        "customer_id": f"cust-{i}",
        "customer_email": f"cust{i}@example.com"
    }
    start = time.perf_counter()
    resp = requests.post(f"{BASE_URL}/shopping-carts", json=payload, timeout=15)
    elapsed = (time.perf_counter() - start) * 1000
    ok = resp.status_code == 201
    record("create_cart", elapsed, ok, resp.status_code)
    if ok:
        cart_ids.append(resp.json()["cart_id"])

# 50 add items
for i in range(50):
    cart_id = cart_ids[i]
    payload = {
        "product_id": f"prod-{i}",
        "name": f"item-{i}",
        "quantity": random.randint(1, 5),
        "unit_price": round(random.uniform(5.0, 100.0), 2)
    }
    start = time.perf_counter()
    resp = requests.post(f"{BASE_URL}/shopping-carts/{cart_id}/items", json=payload, timeout=15)
    elapsed = (time.perf_counter() - start) * 1000
    ok = resp.status_code == 200
    record("add_items", elapsed, ok, resp.status_code)

# 50 get cart
for i in range(50):
    cart_id = cart_ids[i]
    start = time.perf_counter()
    resp = requests.get(f"{BASE_URL}/shopping-carts/{cart_id}", timeout=15)
    elapsed = (time.perf_counter() - start) * 1000
    ok = resp.status_code == 200
    record("get_cart", elapsed, ok, resp.status_code)

with open(RESULTS_FILE, "w", encoding="utf-8") as f:
    json.dump(results, f, indent=2)

print(f"Wrote {len(results)} operations to {RESULTS_FILE}")