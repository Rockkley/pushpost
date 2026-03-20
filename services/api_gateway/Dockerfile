FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/api-gateway ./services/api_gateway/cmd/gateway

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway

EXPOSE 8000

ENTRYPOINT ["/usr/local/bin/api-gateway"]