# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install curl untuk download migrate CLI dan git (swag butuh git untuk fetch dependency)
RUN apk add --no-cache curl git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN go build -ldflags="-s -w" -o main ./inventory_service

# Install migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz \
    | tar -xz && mv migrate /usr/local/bin/migrate

# Stage 2: Run
FROM alpine:3.21

WORKDIR /app

# Copy binary app & migrate CLI dari builder
COPY --from=builder /app/main .
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy folder migration SQL
COPY migrations ./migrations

EXPOSE 9090

# Jalankan aplikasi
CMD ["./main"]
