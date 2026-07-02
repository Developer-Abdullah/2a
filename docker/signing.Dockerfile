FROM debian:bookworm-slim AS zsign-builder
RUN apt-get update && apt-get install -y build-essential libssl-dev zlib1g-dev git unzip zip
RUN git clone https://github.com/zhlynn/zsign.git /zsign-src
WORKDIR /zsign-src
RUN gcc -I./src/third-party/minizip -O3 -Wall -c src/third-party/minizip/ioapi.c \
    && gcc -I./src/third-party/minizip -O3 -Wall -c src/third-party/minizip/unzip.c \
    && gcc -I./src/third-party/minizip -O3 -Wall -c src/third-party/minizip/zip.c \
    && g++ -I./src -I./src/common -I./src/third-party/minizip \
        -O3 -Wall -std=c++11 -o zsign \
        src/*.cpp src/common/*.cpp ioapi.o unzip.o zip.o \
        -lcrypto -lz

FROM golang:alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o signing-worker ./cmd/signing

FROM debian:bookworm-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates zlib1g openssl \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /
COPY --from=zsign-builder /zsign-src/zsign /usr/local/bin/zsign
COPY --from=go-builder /app/signing-worker /signing-worker
ENTRYPOINT ["/signing-worker"]
