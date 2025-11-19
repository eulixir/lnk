FROM golang:1.24.10-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build both migrator and app binaries
RUN go build -o migrator ./cmd/migrator/main.go
RUN go build -o app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

# Copy both binaries
COPY --from=builder /app/migrator .
COPY --from=builder /app/app .

# Default to running the app
CMD ["./app"]