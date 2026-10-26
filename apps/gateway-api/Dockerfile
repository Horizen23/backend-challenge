FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY apps/common ./apps/common
COPY apps/gateway-api ./apps/gateway-api

WORKDIR /src/apps/gateway-api
ENV GOWORK=off
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -tags swagger -o /app/api ./cmd/api

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api .

EXPOSE 8080
ENTRYPOINT ["./api"]
