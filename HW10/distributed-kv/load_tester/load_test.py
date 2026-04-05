import argparse
import csv
import json
import os
import random
import time
import requests
from collections import defaultdict

def parse_args():
    p = argparse.ArgumentParser()
    p.add_argument("--mode", choices=["lf", "leaderless"], required=True)
    p.add_argument("--strategy", choices=["w5r1", "w1r5", "w3r3", "leaderless"], required=True)
    p.add_argument("--ops", type=int, default=500)
    p.add_argument("--read_pct", type=int, required=True)
    p.add_argument("--write_pct", type=int, required=True)
    p.add_argument("--outdir", required=True)
    return p.parse_args()

def main():
    args = parse_args()
    os.makedirs(args.outdir, exist_ok=True)

    if args.mode == "lf":
        leader = "http://localhost:8001"
        read_nodes = [
            "http://localhost:8001",
            "http://localhost:8002",
            "http://localhost:8003",
            "http://localhost:8004",
            "http://localhost:8005",
        ]
        strategy_to_rw = {
            "w5r1": (5, 1),
            "w1r5": (1, 5),
            "w3r3": (3, 3),
        }
        W, R = strategy_to_rw[args.strategy]
    else:
        nodes = [
            "http://localhost:8101",
            "http://localhost:8102",
            "http://localhost:8103",
            "http://localhost:8104",
            "http://localhost:8105",
        ]

    hot_keys = [f"key{i}" for i in range(20)]
    latest_known_version = defaultdict(int)
    last_access_time = {}
    read_latencies = []
    write_latencies = []
    intervals = []
    stale_reads = 0
    rows = []

    for i in range(args.ops):
        do_read = random.randint(1, 100) <= args.read_pct
        key = random.choice(hot_keys)
        now = time.time()

        if key in last_access_time:
            intervals.append(now - last_access_time[key])
        last_access_time[key] = now

        if do_read:
            if args.mode == "lf":
                node = random.choice(read_nodes)
                url = f"{node}/get/{key}?r={R}"
            else:
                node = random.choice(nodes)
                url = f"{node}/get/{key}"

            start = time.perf_counter()
            resp = requests.get(url, timeout=10)
            latency_ms = (time.perf_counter() - start) * 1000
            read_latencies.append(latency_ms)

            stale = False
            version = None
            if resp.status_code == 200:
                body = resp.json()
                version = body["version"]
                if version < latest_known_version[key]:
                    stale_reads += 1
                    stale = True
                else:
                    latest_known_version[key] = max(latest_known_version[key], version)

            rows.append({
                "op_type": "read",
                "key": key,
                "latency_ms": latency_ms,
                "status_code": resp.status_code,
                "version": version,
                "stale": stale,
            })
        else:
            value = f"value-{i}"
            if args.mode == "lf":
                url = f"{leader}/set?w={W}"
            else:
                node = random.choice(nodes)
                url = f"{node}/set"

            start = time.perf_counter()
            resp = requests.post(url, json={"key": key, "value": value}, timeout=20)
            latency_ms = (time.perf_counter() - start) * 1000
            write_latencies.append(latency_ms)

            version = None
            if resp.status_code == 201:
                body = resp.json()
                version = body["version"]
                latest_known_version[key] = max(latest_known_version[key], version)

            rows.append({
                "op_type": "write",
                "key": key,
                "latency_ms": latency_ms,
                "status_code": resp.status_code,
                "version": version,
                "stale": False,
            })

        time.sleep(0.01)

    csv_path = os.path.join(args.outdir, "results.csv")
    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["op_type", "key", "latency_ms", "status_code", "version", "stale"])
        writer.writeheader()
        writer.writerows(rows)

    summary = {
        "mode": args.mode,
        "strategy": args.strategy,
        "ops": args.ops,
        "read_pct": args.read_pct,
        "write_pct": args.write_pct,
        "stale_reads": stale_reads,
        "total_reads": sum(1 for r in rows if r["op_type"] == "read"),
        "total_writes": sum(1 for r in rows if r["op_type"] == "write"),
        "avg_read_latency_ms": sum(read_latencies) / len(read_latencies) if read_latencies else 0,
        "avg_write_latency_ms": sum(write_latencies) / len(write_latencies) if write_latencies else 0,
    }

    with open(os.path.join(args.outdir, "summary.json"), "w", encoding="utf-8") as f:
        json.dump(summary, f, indent=2)

    with open(os.path.join(args.outdir, "intervals.csv"), "w", newline="", encoding="utf-8") as f:
        w = csv.writer(f)
        w.writerow(["interval_seconds"])
        for x in intervals:
            w.writerow([x])

    print("Finished:", summary)

if __name__ == "__main__":
    main()