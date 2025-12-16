FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o node-manager

# ---- runtime ----
FROM ubuntu:22.04

RUN apt-get update && \
    apt-get install -y \
      ca-certificates \
      curl \
      tar && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /tmp

RUN curl --fail -LS \
    https://github.com/liqotech/liqo/releases/download/v1.0.0/liqoctl-linux-amd64.tar.gz \
    | tar -xz && \
    install -o root -g root -m 0755 liqoctl /usr/local/bin/liqoctl && \
    rm -rf /tmp/*

WORKDIR /app
COPY --from=builder /app/node-manager /app/node-manager

EXPOSE 9009
ENTRYPOINT ["/app/node-manager"]


