# Reliability Copilot Final Experiments Report

## Experiment 1 — Bottleneck Identification

### Purpose
The purpose of this experiment was to identify the first major bottleneck in the system as load increased.

### Tradeoff Explored
The main tradeoff explored was throughput versus latency under higher request volume.

### Limitations
This experiment used a simplified local deployment and lightweight classification logic, so it does not fully capture cloud-network effects or more expensive inference paths.

### Results
Include:
- small load chart
- medium load chart
- heavy load chart
- p50 / p95 summary
- throughput summary

### Analysis
Discuss:
- where latency started rising sharply
- whether the API, worker, or storage path looked like the bottleneck
- what evidence supports that conclusion
- limitations of the evidence

## Experiment 2 — Scaling Behavior

### Purpose
The purpose of this experiment was to test whether increasing worker capacity improved system stability and throughput.

### Tradeoff Explored
The main tradeoff explored was added worker capacity versus latency and queue pressure.

### Limitations
This was simulated scaling through worker count changes rather than full production auto scaling.

### Results
Include:
- worker 1 result
- worker 4 result
- worker 8 result
- comparison table

### Analysis
Discuss:
- whether p95 latency improved
- whether throughput improved
- whether scaling helped enough to justify extra complexity
- limitations of the result

## Experiment 3 — Failure Injection and Recovery

### Purpose
The purpose of this experiment was to compare an unprotected design with a protected design under artificial failure conditions.

### Tradeoff Explored
The main tradeoff explored was simplicity versus resilience.

### Limitations
The failures were simulated rather than coming from a real cloud dependency.

### Results
Include:
- unprotected run charts
- protected run charts
- failure rate / latency comparison

### Analysis
Discuss:
- whether the unprotected version degraded more severely
- whether fallback behavior helped
- what evidence supports your conclusion
- what is still missing from the analysis

## Final Conclusion
Summarize:
- what broke first
- whether scaling helped
- whether resilience protections helped
- what you would improve next