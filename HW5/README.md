CS6650 – HW5: Scalable Online Store + ECS Deployment
Overview

This project contains:

Part II A simple online store Product API (Go, net/http)

Part III A Gin-based Albums API deployed using ECS + ECR + Terraform

Part IV Load testing using Locust

Infrastructure as Code - Managed fully with Terraform

My full report and screenshots are included in:
public/CS6650HW5WNReport.pdf

📁 Project Structure
HW5/
├── CS6650_2b_demo/ → ECS + Terraform deployment (Part III)
│ ├── src/ → Gin albums API
│ ├── terraform/ → Infrastructure code
│ ├── tests/ → Locust load tests
│ └── README.md
│
├── online-store-product-api/ → Product API (Part II)
│ └── src/
│ ├── main.go
│ └── Dockerfile
│
├── public/CS6650HW5WNReport.pdf
└── .gitignore

Product API (Part II)

Located in:

online-store-product-api/src/main.go

This supports:

GET /products/{productId}

POST /products/{productId}/details

GET /health

Example Response Codes
200 – Product Found
GET /products/12345

204 – Product Details Added
POST /products/12345/details

400 – Invalid Input

If body fails validation.

404 – Product Not Found

If product ID does not exist.

500 – Internal Server Error

Triggered manually via panic during testing (see report screenshots in PDF

CS6650HW5WNReport

).

How To Deploy Infrastructure On A New Machine

These steps allow anyone in my group to deploy everything easily.

✅ 1. Install Requirements

Docker Desktop

Terraform

AWS CLI

✅ 2. Configure AWS
aws configure

Enter:

Access Key

Secret Key

Region (us-east-1)

Default output: json

✅ 3. Deploy Infrastructure

Navigate to:

cd HW5/CS6650_2b_demo/terraform

Initialize Terraform:

terraform init

Apply infrastructure:

terraform apply -auto-approve

Terraform will:

Create ECR repository

Build Docker image

Push image to ECR

Create ECS cluster

Create ECS service

Create networking + security group

Create CloudWatch logs

✅ 4. Get Public IP of ECS Task
aws ecs list-tasks \
 --cluster CS6650L2-cluster

aws ecs describe-tasks \
 --cluster CS6650L2-cluster \
 --tasks <task-id> \
 --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
 --output text

Then:

aws ec2 describe-network-interfaces \
 --network-interface-ids <eni-id> \
 --query "NetworkInterfaces[0].Association.PublicIp" \
 --output text

You can now send requests to:

http://<public-ip>:8080

API Endpoints (Albums API – Part III)

Located in:

CS6650_2b_demo/src/main.go

GET all albums
GET /albums

Example:

curl http://<public-ip>:8080/albums

Response:

200 OK

GET album by ID
GET /albums/:id

Example:

curl http://<public-ip>:8080/albums/1

Responses:

200 OK

404 Not Found

POST album
POST /albums

Example:

curl -X POST http://<public-ip>:8080/albums \
 -H "Content-Type: application/json" \
 -d '{"id":"4","title":"Test","artist":"Me","price":10.0}'

Responses:

201 Created

Load Testing (Locust)

Located in:

CS6650_2b_demo/tests/

Run normal HttpUser test:

locust -f tests/locustfile.py --host "http://<public-ip>:8080"

Run FastHttpUser test:

locust -f tests/locustfile_fast.py --host "http://<public-ip>:8080"

Test Scenarios I Ran:

50 users, spawn 5

200 users, spawn 10

500 users, spawn 20

FastHttpUser comparison

Results and graphs are shown in:
public/CS6650HW5WNReport.pdf

CS6650HW5WNReport

Design Notes

Reads (GETs) dominate in real-world systems

I used maps (hashmaps) for fast lookup

Infrastructure is modular (network, ecr, ecs, logging)

ECS allows horizontal scaling via ecs_count

Terraform makes deployment reproducible and safe

Infrastructure Code Location
CS6650_2b_demo/terraform/

Modules:

network/

ecr/

ecs/

logging/

Main files:

main.tf

provider.tf

variables.tf

outputs.tf

Dockerfiles

Product API → online-store-product-api/src/Dockerfile

Albums API → CS6650_2b_demo/src/Dockerfile

.gitignore

The repository excludes:

terraform.tfstate

terraform.tfstate.backup

.terraform/

binaries

.env

.tfvars

AWS credentials

large files

This keeps repo clean and secure.
