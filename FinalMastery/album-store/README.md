# Album Store

This service implements the CS 6650 ChaosArena Album Store contract.

## Stack

- Go API
- Python background worker
- SQLite
- Docker Compose for local development
- Terraform + EC2 for public deployment

## Endpoints

- GET /health
- PUT /albums/:album_id
- GET /albums/:album_id
- GET /albums
- POST /albums/:album_id/photos
- GET /albums/:album_id/photos/:photo_id
- DELETE /albums/:album_id/photos/:photo_id

## Local Run

1. Set `PUBLIC_BASE_URL=http://localhost` in `.env`
2. Run `docker compose up --build -d`
3. Run `python tests/smoke_test.py`

## Terraform Deployment

1. Configure AWS CLI with lab credentials
2. Set `terraform/terraform.tfvars`
3. Run `terraform init`
4. Run `terraform apply -auto-approve`
5. Use the output `base_url`
