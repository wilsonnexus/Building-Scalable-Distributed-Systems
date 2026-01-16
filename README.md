## web-service-gin (Go + Gin REST API)

A small RESTful API written in Go using the Gin framework. It exposes a tiny in-memory “albums” service with three endpoints:

- GET /albums — list all albums
- GET /albums/:id — get an album by id
- POST /albums — add a new album (with basic validation + duplicate-id check)

This project was built as a Week 1 intro exercise for CS 6650 (Building Scalable Distributed Systems) to practice Go tooling, REST API structure, lightweight error handling, and simple tests.

---

## PREREQUISITES

- Go installed (Go 1.16+ recommended)
- Internet access the first time to download dependencies (go get .)
- curl (PowerShell note below)

---

## RUN LOCALLY (WINDOWS)

From the project folder:

go get .
go run .

The server listens on:
http://localhost:8080

---

## BASIC API CHECKS (GET)

curl http://localhost:8080/albums
curl http://localhost:8080/albums/2

---

## WINDOWS POWERSHELL (POST)

PowerShell has quoting differences, so use curl.exe.

Valid POST:

curl.exe --% -i -H "Content-Type: application/json" -X POST http://localhost:8080/albums -d "{\"id\":\"4\",\"title\":\"The Modern Sound of Betty Carter\",\"artist\":\"Betty Carter\",\"price\":49.99}"

Invalid JSON (should return 400):

curl.exe -i -H "Content-Type: application/json" -X POST http://localhost:8080/albums -d "not-json"

---

## RUN ON GOOGLE CLOUD PLATFORM (GCP)

In Google Cloud Shell, open the project folder and run:

go get .
go run .

---

## TEST ENDPOINTS (CLOUD SHELL / BASH)

curl http://localhost:8080/albums
curl http://localhost:8080/albums/2

Valid POST (bash single quotes):

curl -i -H "Content-Type: application/json" \
 -X POST http://localhost:8080/albums \
 -d '{"id":"4","title":"The Modern Sound of Betty Carter","artist":"Betty Carter","price":49.99}'

Invalid JSON (should return 400):

curl -i -H "Content-Type: application/json" -X POST http://localhost:8080/albums -d 'not-json'

---

## RUN TESTS

This project includes basic endpoint tests using Go’s net/http/httptest package:

go test -v

---

## NOTES

- Data is stored in-memory (albums slice). Restarting the server resets the data.
- POST requests include basic validation and duplicate-id checks (409 Conflict).

