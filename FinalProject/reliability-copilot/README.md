# Reliability Copilot

Reliability Copilot is a small but realistic AWS-based distributed systems project that analyzes simulated CI/CD and build/test failures.

The system accepts failure events, processes them asynchronously, and returns:

- a failure category
- a short recommended action
- a risk score for repeat failure

## Why I built it

I built this project because reliability incidents are hard to triage when logs are noisy, repetitive, or unclear. I wanted to build something that combines distributed systems design, observability, queue-based processing, resilience under failure, and lightweight AI-assisted analysis in one realistic platform.

I also wanted this project to go beyond just having AI classify logs. The bigger goal was to design and evaluate the full system around that AI component and show how it behaves under load, scaling changes, and failure conditions.

## Architecture

The final version of the project is AWS-based and Terraform-managed. It includes:

- Go API service on ECS Fargate behind an Application Load Balancer
- Amazon SQS for async job buffering and queue-based decoupling
- Go worker service on ECS Fargate for background processing
- Amazon DynamoDB for event and result storage
- CloudWatch for logs and system metrics
- Terraform for infrastructure provisioning
- Locust for load testing and experiments

## Core Experiments

This project was organized around three main experiments:

1. **Bottleneck Identification**  
   Increase load and identify what breaks first or slows down first.

2. **Scaling Behavior**  
   Test whether adding worker-side capacity improves throughput and latency stability.

3. **Failure Injection and Recovery**  
   Compare an unprotected version and a protected version under artificial delays and failures.

## What the project shows

The project is meant to answer three main systems questions:

- What becomes the first bottleneck as load increases?
- Does scaling actually help the system stay responsive?
- Do resilience patterns help the system degrade more gracefully under failure?

## Repo Contents

- `cmd/api` — Go API service
- `cmd/worker` — Go worker service
- `internal/classifier` — lightweight failure classification logic
- `internal/model` — shared event/result types
- `internal/store` — storage and AWS integration logic
- `terraform` — AWS infrastructure configuration
- `tests` — Locust load tests and plotting scripts
- `scripts` — build, deploy, and experiment run scripts
- `docs` — project management, report, lessons learned, and video script
- `results` — experiment outputs and charts

## Code and Project Management Links

- **Repo:** https://github.com/wilsonnexus/Building-Scalable-Distributed-Systems/tree/main/FinalProject/reliability-copilot
- **GitHub Project Board:** https://github.com/users/wilsonnexus/projects/2/views/1

## Final note

I worked on this project solo, so I kept the scope realistic while still making sure it reflected the original proposal: AWS deployment, queue-based distributed design, observability, and experiments that show system tradeoffs under load and failure.
