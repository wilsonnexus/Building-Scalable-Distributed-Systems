# HW8 - Shopping Cart API with MySQL (RDS) and DynamoDB

## Overview

This project implements a cloud-based shopping cart API on AWS and compares two database backends for the same application workload:

- **MySQL on Amazon RDS** for a relational implementation
- **Amazon DynamoDB** for a NoSQL implementation

The goal of the assignment is to build the same shopping cart API twice, run the same test workload against both versions, collect performance results, and compare the trade-offs between relational and NoSQL approaches.

The API supports:

- `POST /shopping-carts`
- `GET /shopping-carts/{id}`
- `POST /shopping-carts/{id}/items`
- `GET /health`

---

## Assignment Goals

This homework is divided into three main parts:

### Step I - MySQL Integration for Relational Data

Build the shopping cart API using:

- **Amazon RDS MySQL**
- proper relational schema design
- indexes and connection pooling
- ECS deployment and CloudWatch monitoring
- a 150-operation performance test

### Step II - DynamoDB Integration for NoSQL Data

Build the same shopping cart API using:

- **Amazon DynamoDB**
- a NoSQL cart design
- ECS deployment and CloudWatch monitoring
- eventual consistency investigation
- the same 150-operation performance test

### Step III - Database Comparison and Analysis

Compare MySQL and DynamoDB using:

- `mysql_test_results.json`
- `dynamodb_test_results.json`
- `combined_results.json`

The comparison includes:

- average and percentile response times
- operation-specific performance
- consistency trade-offs
- resource efficiency
- architecture recommendations
- polyglot persistence strategy

---

## Architecture

### Application Layer

The application is a Go service deployed on **ECS Fargate** behind an **Application Load Balancer (ALB)**.

It exposes the same shopping cart API regardless of database backend.

### MySQL Mode

Client → ALB → ECS Go API → Amazon RDS MySQL

This version uses a relational schema with separate tables for carts and cart items.

### DynamoDB Mode

Client → ALB → ECS Go API → Amazon DynamoDB

This version uses a NoSQL design with cart data stored by `cart_id`, including embedded items.

---

## Infrastructure Requirements Implemented

- **VPC CIDR:** `10.0.0.0/16`
- **Public Subnets:** `10.0.1.0/24`, `10.0.2.0/24`
- **Private Subnets:** `10.0.10.0/24`, `10.0.11.0/24`
- **ALB in public subnets**
- **ECS service in private subnets**
- **NAT Gateway for outbound access**
- **Amazon RDS MySQL** in private subnets
- **Amazon DynamoDB** table for NoSQL cart storage

### ECS Task Settings

The ECS task uses:

- CPU: `256`
- Memory: `512`
- Health check: `GET /health`

---

## API Endpoints

### Create Shopping Cart

```http
POST /shopping-carts
```
