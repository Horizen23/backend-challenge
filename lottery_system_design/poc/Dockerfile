# Build Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/lottery-poc ./main.go

# Runtime Stage
FROM alpine:3.21

WORKDIR /app
COPY --from=builder /app/lottery-poc /app/lottery-poc

ENTRYPOINT ["/app/lottery-poc"]
