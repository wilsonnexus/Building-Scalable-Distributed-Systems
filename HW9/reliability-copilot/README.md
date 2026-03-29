# Reliability Copilot

## Overview

Reliability Copilot is a small but realistic AWS-based platform for analyzing simulated CI/CD and build/test failures.

It accepts failure events, classifies them into categories such as timeout, flaky test, infrastructure issue, or dependency failure, and returns:

- a failure category
- a short recommended action
- a risk score for repeat failure

This project is being built as a distributed systems capstone with a focus on:

- scalable infrastructure
- reliability and failure handling
- observability
- lightweight AI / LLM-assisted analysis

## Initial Scope

For the checkpoint submission, this repo includes:

- a starter API service
- a simple load test
- chart generation scripts
- a project plan
- a report outline
- an elevator pitch script

## Planned Architecture

- API Service in Go
- Queue-based decoupling
- Worker service
- Storage for results
- CloudWatch metrics/logs
- experiments on bottlenecks, scaling, and failure resilience