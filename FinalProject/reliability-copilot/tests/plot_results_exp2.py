import sys
import os
import pandas as pd
import matplotlib.pyplot as plt

if len(sys.argv) != 2:
    print("Usage: python tests/plot_results_exp2.py results/exp2")
    raise SystemExit(1)

folder = sys.argv[1]

configs = [
    ("autoscale_aws", "AWS Autoscale Run"),
]

loaded = []

def find_column(df, candidates):
    for c in candidates:
        if c in df.columns:
            return c
    return None

for prefix, label in configs:
    path = os.path.join(folder, f"{prefix}_stats_history.csv")
    if os.path.exists(path):
        df = pd.read_csv(path)
        print(f"\nLoaded {path}")
        print("Columns:", list(df.columns))
        loaded.append((prefix, label, df))
    else:
        print(f"Warning: file not found: {path}")

if not loaded:
    raise FileNotFoundError(f"No *_stats_history.csv files found in {folder}")

timestamp_candidates = ["Timestamp", "Time", "timestamp"]
rps_candidates = ["Total RPS", "Requests/s", "Current RPS"]
avg_latency_candidates = ["Average Response Time", "Avg Response Time", "Average response time"]
p95_candidates = ["95%", "95th percentile", "P95", "95th"]

sample_df = loaded[0][2]
timestamp_col = find_column(sample_df, timestamp_candidates)

if timestamp_col is None:
    raise KeyError(f"Could not find a timestamp column. Available columns: {list(sample_df.columns)}")

prefix, label, df = loaded[0]

rps_col = find_column(df, rps_candidates)
if rps_col is not None:
    plt.figure()
    plt.plot(df[timestamp_col], df[rps_col], label=label)
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Requests/sec")
    plt.title("AWS Experiment 2 Throughput")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "throughput_comparison.png"))

avg_col = find_column(df, avg_latency_candidates)
if avg_col is not None:
    plt.figure()
    plt.plot(df[timestamp_col], df[avg_col], label=label)
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Milliseconds")
    plt.title("AWS Experiment 2 Average Latency")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "latency_avg_comparison.png"))

p95_col = find_column(df, p95_candidates)
if p95_col is not None:
    plt.figure()
    plt.plot(df[timestamp_col], df[p95_col], label=label)
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Milliseconds")
    plt.title("AWS Experiment 2 P95 Latency")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "latency_p95_comparison.png"))

print("\nSaved charts to", folder)