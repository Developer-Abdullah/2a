FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o analytics-worker ./cmd/analytics

FROM alpine:latest
WORKDIR /
COPY --from=builder /app/analytics-worker /analytics-worker
ENTRYPOINT ["/analytics-worker"]
