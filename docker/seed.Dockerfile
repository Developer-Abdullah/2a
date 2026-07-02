FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o seed ./cmd/seed

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /app/seed /seed
ENTRYPOINT ["/seed"]
