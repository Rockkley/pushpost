FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/notification-service ./services/notification_service/cmd/notification

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/notification-service /usr/local/bin/notification-service

EXPOSE 8084


ENTRYPOINT ["/usr/local/bin/notification-service"]
