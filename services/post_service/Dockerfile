FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /out/post-service ./services/post_service/cmd/post

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /service

COPY --from=builder /out/post-service /usr/local/bin/post-service

EXPOSE 8085
EXPOSE 9082

ENTRYPOINT ["/usr/local/bin/post-service"]
