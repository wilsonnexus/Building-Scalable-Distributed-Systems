import threading
import time
import requests

NODES = [
    "http://localhost:8101",
    "http://localhost:8102",
    "http://localhost:8103",
    "http://localhost:8104",
    "http://localhost:8105",
]

def test_leaderless_after_ack_reads_are_consistent():
    coordinator = NODES[0]
    other = NODES[1]
    key = "gamma"
    value = "three"

    r = requests.post(f"{coordinator}/set", json={"key": key, "value": value}, timeout=15)
    assert r.status_code == 201

    r1 = requests.get(f"{coordinator}/get/{key}", timeout=5)
    assert r1.status_code == 200
    assert r1.json()["value"] == value

    r2 = requests.get(f"{other}/get/{key}", timeout=5)
    assert r2.status_code == 200
    assert r2.json()["value"] == value

def test_leaderless_inconsistency_window():
    coordinator = NODES[0]
    others = NODES[1:]
    key = "delta"
    value = "four"

    def do_write():
        requests.post(f"{coordinator}/set", json={"key": key, "value": value}, timeout=15)

    t = threading.Thread(target=do_write)
    t.start()

    time.sleep(0.15)

    saw_inconsistency = False
    for node in others:
        r = requests.get(f"{node}/get/{key}", timeout=5)
        if r.status_code == 404:
            saw_inconsistency = True
            break
        elif r.status_code == 200 and r.json()["value"] != value:
            saw_inconsistency = True
            break

    t.join()
    assert saw_inconsistency