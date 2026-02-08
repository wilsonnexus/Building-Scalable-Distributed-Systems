import matplotlib.pyplot as plt

# Your measured times (seconds)
times_s = {
    "split": 0.247902,
    "map0":  0.148496,
    "map1":  0.150162,
    "map2":  0.127319,
    "reduce":0.275870,
}

labels = list(times_s.keys())
times_ms = [times_s[k] * 1000 for k in labels]  # convert to ms

plt.figure()
plt.bar(labels, times_ms)
plt.ylabel("Time (ms)")
plt.title("MapReduce Lab: Split / Map / Reduce Latency (single run)")
plt.tight_layout()
plt.savefig("latency_bar.png", dpi=200)
plt.show()
print("Saved: latency_bar.png")
