FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/auth-service ./services/auth_service/cmd/auth

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/auth-service /usr/local/bin/auth-service
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/auth-service"]