import threading
import time
import requests

LEADER = "http://localhost:8001"
FOLLOWERS = [
    "http://localhost:8002",
    "http://localhost:8003",
    "http://localhost:8004",
    "http://localhost:8005",
]

def test_leader_write_then_consistent_reads():
    key = "alpha"
    value = "one"

    r = requests.post(f"{LEADER}/set?w=5", json={"key": key, "value": value}, timeout=10)
    assert r.status_code == 201

    r1 = requests.get(f"{LEADER}/get/{key}?r=1", timeout=5)
    assert r1.status_code == 200
    assert r1.json()["value"] == value

    r2 = requests.get(f"{FOLLOWERS[0]}/get/{key}?r=1", timeout=5)
    assert r2.status_code == 200
    assert r2.json()["value"] == value

def test_leader_follower_inconsistency_window_with_local_read():
    key = "beta"
    value = "two"

    def do_write():
        requests.post(f"{LEADER}/set?w=5", json={"key": key, "value": value}, timeout=15)

    t = threading.Thread(target=do_write)
    t.start()

    time.sleep(0.15)

    saw_inconsistency = False
    for follower in FOLLOWERS:
        r = requests.get(f"{follower}/local_read/{key}", timeout=5)
        if r.status_code == 404:
            saw_inconsistency = True
            break
        elif r.status_code == 200 and r.json()["value"] != value:
            saw_inconsistency = True
            break

    t.join()
    assert saw_inconsistency