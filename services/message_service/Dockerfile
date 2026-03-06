FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/message-service ./services/message_service/cmd/message

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/message-service /usr/local/bin/message-service
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/message-service"]