FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/user-service ./services/user_service/cmd/user

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/user-service /usr/local/bin/user-service
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/user-service"]