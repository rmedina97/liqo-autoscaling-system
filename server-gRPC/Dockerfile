FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server-grpc .


COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o server-grpc .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/server-grpc .

EXPOSE 9007

ENTRYPOINT ["./server-grpc"]
