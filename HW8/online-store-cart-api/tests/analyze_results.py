import json
import statistics

def percentile(values, p):
    if not values:
        return 0
    values = sorted(values)
    k = (len(values) - 1) * (p / 100)
    f = int(k)
    c = min(f + 1, len(values) - 1)
    if f == c:
        return values[int(k)]
    return values[f] + (values[c] - values[f]) * (k - f)

def summarize(data):
    times = [x["response_time"] for x in data]
    success_rate = (sum(1 for x in data if x["success"]) / len(data)) * 100
    return {
        "avg": round(statistics.mean(times), 2),
        "p50": round(percentile(times, 50), 2),
        "p95": round(percentile(times, 95), 2),
        "p99": round(percentile(times, 99), 2),
        "success_rate": round(success_rate, 2),
        "total": len(data),
    }

def summarize_op(data, op):
    subset = [x["response_time"] for x in data if x["operation"] == op]
    return round(statistics.mean(subset), 2)

with open("combined_results.json", "r", encoding="utf-8") as f:
    combined = json.load(f)

mysql = combined["mysql"]
ddb = combined["dynamodb"]

mysql_summary = summarize(mysql)
ddb_summary = summarize(ddb)

print("MYSQL", mysql_summary)
print("DYNAMODB", ddb_summary)

for op in ["create_cart", "add_items", "get_cart"]:
    print(op, "mysql", summarize_op(mysql, op), "dynamodb", summarize_op(ddb, op))