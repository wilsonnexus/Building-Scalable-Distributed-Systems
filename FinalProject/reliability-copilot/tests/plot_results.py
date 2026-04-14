import sys
import os
import pandas as pd
import matplotlib.pyplot as plt

if len(sys.argv) != 2:
    print("Usage: python tests/plot_results.py results/exp1")
    raise SystemExit(1)

folder = sys.argv[1]

configs = [
    ("small", "Small Load"),
    ("medium", "Medium Load"),
    ("heavy", "Heavy Load"),
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

# Common Locust column-name variations
timestamp_candidates = ["Timestamp", "Time", "timestamp"]
rps_candidates = ["Total RPS", "Requests/s", "Current RPS", "Total Request Count", "Request Count"]
avg_latency_candidates = ["Average Response Time", "Avg Response Time", "Average response time"]
p95_candidates = ["95%", "95th percentile", "P95", "95th"]

# Use the first loaded file to discover the timestamp column
sample_df = loaded[0][2]
timestamp_col = find_column(sample_df, timestamp_candidates)

if timestamp_col is None:
    raise KeyError(f"Could not find a timestamp column. Available columns: {list(sample_df.columns)}")

# Throughput comparison chart
plt.figure()
plotted_rps = False
for prefix, label, df in loaded:
    rps_col = find_column(df, rps_candidates)
    if rps_col is not None:
        plt.plot(df[timestamp_col], df[rps_col], label=label)
        plotted_rps = True

if plotted_rps:
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Requests/sec")
    plt.title("Experiment 1 Throughput Comparison")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "throughput_comparison.png"))
else:
    print("No RPS-style column found, skipping throughput chart.")
    plt.close()

# Average latency comparison chart
plt.figure()
plotted_avg = False
for prefix, label, df in loaded:
    avg_col = find_column(df, avg_latency_candidates)
    if avg_col is not None:
        plt.plot(df[timestamp_col], df[avg_col], label=label)
        plotted_avg = True

if plotted_avg:
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Milliseconds")
    plt.title("Experiment 1 Average Latency Comparison")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "latency_avg_comparison.png"))
else:
    print("No average latency column found, skipping average latency chart.")
    plt.close()

# P95 comparison chart
plt.figure()
plotted_p95 = False
for prefix, label, df in loaded:
    p95_col = find_column(df, p95_candidates)
    if p95_col is not None:
        plt.plot(df[timestamp_col], df[p95_col], label=label)
        plotted_p95 = True

if plotted_p95:
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Milliseconds")
    plt.title("Experiment 1 P95 Latency Comparison")
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(folder, "latency_p95_comparison.png"))
else:
    print("No p95-style column found, skipping p95 chart.")
    plt.close()

print("\nSaved charts to", folder)