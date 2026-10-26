FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY apps/common ./apps/common
COPY apps/user-service ./apps/user-service

WORKDIR /src/apps/user-service
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/grpc-service ./cmd/grpc

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/grpc-service .

EXPOSE 50051
ENTRYPOINT ["./grpc-service"]
