FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o notifier-worker ./cmd/notifier

FROM alpine:latest
WORKDIR /
COPY --from=builder /app/notifier-worker /notifier-worker
ENTRYPOINT ["/notifier-worker"]
