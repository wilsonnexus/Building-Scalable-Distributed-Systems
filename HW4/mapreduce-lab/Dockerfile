# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app

# Leverage cache
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Which service to build: splitter | mapper | reducer
ARG SERVICE=splitter
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./${SERVICE}

# Runtime stage
FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 8080
ENV ADDR=:8080
CMD ["/app/app"]
