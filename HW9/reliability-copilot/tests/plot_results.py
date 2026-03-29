import pandas as pd
import matplotlib.pyplot as plt

history = pd.read_csv("results/initial_stats_history.csv")

print("Columns found:")
print(list(history.columns))

# Try to detect the timestamp column
timestamp_col = None
for col in ["Timestamp", "timestamp"]:
    if col in history.columns:
        timestamp_col = col
        break

if timestamp_col is None:
    raise KeyError("Could not find a Timestamp column in initial_stats_history.csv")

# Try to detect RPS column
rps_col = None
for col in ["Total RPS", "Requests/s", "Total Request Count/s", "Total Average Response Time"]:
    if col in history.columns:
        rps_col = col
        break

# Try to detect average response time column
avg_resp_col = None
for col in ["Average Response Time", "Avg Response Time", "Total Average Response Time"]:
    if col in history.columns:
        avg_resp_col = col
        break

# Try to detect p95 column
p95_col = None
for col in ["95%", "95th Percentile", "Total 95%"]:
    if col in history.columns:
        p95_col = col
        break

# Throughput chart
if rps_col is not None:
    plt.figure()
    plt.plot(history[timestamp_col], history[rps_col])
    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Requests/sec")
    plt.title("Initial Reliability Copilot Throughput")
    plt.tight_layout()
    plt.savefig("results/throughput_chart.png")
    print(f"Saved throughput_chart.png using column: {rps_col}")
else:
    print("No throughput/RPS column found, skipping throughput chart.")

# Latency chart
if avg_resp_col is not None:
    plt.figure()
    plt.plot(history[timestamp_col], history[avg_resp_col], label="Avg Response Time")

    if p95_col is not None:
        plt.plot(history[timestamp_col], history[p95_col], label="P95 Response Time")

    plt.xticks(rotation=45, ha="right")
    plt.xlabel("Time")
    plt.ylabel("Milliseconds")
    plt.title("Initial Reliability Copilot Latency")
    plt.legend()
    plt.tight_layout()
    plt.savefig("results/latency_chart.png")
    print(f"Saved latency_chart.png using columns: {avg_resp_col}" + (f", {p95_col}" if p95_col else ""))
else:
    print("No average response time column found, skipping latency chart.")