FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o core-api ./cmd/api

FROM alpine:latest
WORKDIR /
COPY --from=builder /app/core-api /core-api
EXPOSE 8080
ENTRYPOINT ["/core-api"]
