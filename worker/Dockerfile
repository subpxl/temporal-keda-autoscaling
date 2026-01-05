FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy go.mod and go.sum from root of /app
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build worker
RUN CGO_ENABLED=0 GOOS=linux go build -o worker ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/worker .
CMD ["./worker"]
