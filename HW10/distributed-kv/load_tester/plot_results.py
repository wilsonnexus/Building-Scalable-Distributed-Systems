import argparse
import pandas as pd
import matplotlib.pyplot as plt
import os

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--indir", required=True)
    args = p.parse_args()

    results = pd.read_csv(os.path.join(args.indir, "results.csv"))
    intervals = pd.read_csv(os.path.join(args.indir, "intervals.csv"))

    reads = results[results["op_type"] == "read"]
    writes = results[results["op_type"] == "write"]

    plt.figure()
    plt.hist(reads["latency_ms"], bins=30)
    plt.xlabel("Read Latency (ms)")
    plt.ylabel("Count")
    plt.title("Distribution of Read Latency")
    plt.tight_layout()
    plt.savefig(os.path.join(args.indir, "read_latency_hist.png"))

    plt.figure()
    plt.hist(writes["latency_ms"], bins=30)
    plt.xlabel("Write Latency (ms)")
    plt.ylabel("Count")
    plt.title("Distribution of Write Latency")
    plt.tight_layout()
    plt.savefig(os.path.join(args.indir, "write_latency_hist.png"))

    plt.figure()
    plt.hist(intervals["interval_seconds"], bins=30)
    plt.xlabel("Time Interval Between Accesses to Same Key (s)")
    plt.ylabel("Count")
    plt.title("Distribution of Read/Write Time Intervals on Same Key")
    plt.tight_layout()
    plt.savefig(os.path.join(args.indir, "interval_hist.png"))

    print("Graphs saved in", args.indir)

if __name__ == "__main__":
    main()