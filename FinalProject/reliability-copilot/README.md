# Reliability Copilot

Reliability Copilot is a small but realistic distributed-systems project that analyzes simulated CI/CD and build/test failures.

The system accepts failure events, processes them asynchronously, and returns:

- a failure category
- a short recommended action
- a risk score for repeat failure

## Why I built it

I built this project because reliability incidents are hard to triage when logs are noisy, repetitive, or unclear. I wanted to build something that combines distributed systems design, observability, queue-based processing, and lightweight AI-assisted analysis in one realistic platform.

## Architecture

- Go API service
- Go worker service
- SQLite-backed queue/result store
- Docker Compose for local orchestration
- Locust-based experiment scripts
- Planned AWS deployment if needed for demo

## Core Experiments

1. Bottleneck Identification  
2. Scaling Behavior  
3. Failure Injection and Recovery  

## Repo Contents

- `cmd/api` — API service
- `cmd/worker` — background worker
- `internal/classifier` — lightweight classification logic
- `internal/store` — SQLite storage layer
- `tests` — Locust load tests and plotting
- `scripts` — experiment run scripts
- `docs` — project management, experiment report, lessons learned, and video script
- `results` — experiment outputs and charts