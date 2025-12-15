# Stage 1 — Build binary
FROM golang:1.24-alpine AS builder

WORKDIR /app
# Cache dependencies,otherwise they will be downloaded at each build
COPY go.mod go.sum ./
RUN go mod download

# Copy all the source files
COPY . .
# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o discovery-server

# Stage 2 — No shell or package manager, just the binary
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/discovery-server /app/discovery-server
COPY certificates /app/certificates

# Rest API Port
EXPOSE 9010
ENTRYPOINT ["/app/discovery-server"]
