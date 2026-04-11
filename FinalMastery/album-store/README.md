# Album Store

This service implements the CS 6650 ChaosArena Album Store contract.

## Stack

- Go API
- Python background worker
- SQLite
- Docker Compose
- EC2 deployment

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

## Deployment

Deploy to a public EC2 instance and set `PUBLIC_BASE_URL` to that host.
