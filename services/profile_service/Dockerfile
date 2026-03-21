FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/profile-service ./services/profile_service/cmd/profile

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/profile-service /usr/local/bin/profile-service

ENTRYPOINT ["/usr/local/bin/profile-service"]
