FROM golang:alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o admin-api ./cmd/admin

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /app/admin-api /admin-api
EXPOSE 8081
ENTRYPOINT ["/admin-api"]
