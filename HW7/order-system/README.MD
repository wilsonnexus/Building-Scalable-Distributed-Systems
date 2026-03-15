# HW7 - Order Processing: Sync vs Async with ECS, SNS, SQS, and Lambda

## Overview

This project demonstrates why synchronous request processing fails under burst traffic and how an asynchronous event-driven architecture improves acceptance rate.

The system includes:

- **Order API (ECS Fargate)**  
  Exposes:
  - `POST /orders/sync`
  - `POST /orders/async`
  - `GET /health`

- **Order Processor (ECS Fargate)**  
  Continuously polls SQS and processes orders in the background.

- **Amazon SNS**  
  Topic used by the async API to publish accepted orders.

- **Amazon SQS**  
  Queue subscribed to the SNS topic, used to buffer background work.

- **AWS Lambda (Part III)**  
  Optional serverless replacement for ECS workers, subject to lab IAM permissions.

---

## Architecture

### Part II - Synchronous vs Asynchronous Order Processing

### Synchronous flow

Customer → API → Payment Verification (3s) → Response

### Asynchronous flow

Customer → API → SNS → SQS → ECS Workers  
Customer ← immediate `202 Accepted`

The sync version simulates a payment bottleneck with a buffered Go channel so that only **1 payment can be processed at a time**, even under concurrency.

---

## Infrastructure Requirements Implemented

- **VPC CIDR:** `10.0.0.0/16`
- **Public Subnets:** `10.0.1.0/24`, `10.0.2.0/24`
- **Private Subnets:** `10.0.10.0/24`, `10.0.11.0/24`
- **ALB in public subnets**
- **ECS services in private subnets**
- **NAT Gateway for outbound access**
- **SNS topic:** `order-processing-events`
- **SQS queue:** `order-processing-queue`
- **Queue visibility timeout:** 30s
- **Message retention:** 4 days
- **Receive wait time:** 20s

### ECS task settings

All ECS tasks use:

- CPU: `256`
- Memory: `512`
- Health check: `GET /health`

---

## Project Structure

```text
order-system/
  src/
    main.go
    Dockerfile
    lambda_main.go
    go.mod
    go.sum
    lambda.zip
  tests/
    locust_sync.py
    locust_async.py
  terraform/
    main.tf
    variables.tf
    sns_sqs.tf
    ecs_services.tf
    outputs.tf
    lambda.tf
```
