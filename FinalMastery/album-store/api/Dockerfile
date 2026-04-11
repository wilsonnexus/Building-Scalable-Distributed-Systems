FROM golang:1.22-bullseye AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN apt-get update && apt-get install -y gcc g++ libc6-dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o server main.go

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]