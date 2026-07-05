FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy source first, then run go mod tidy to resolve and download all dependencies
COPY go.mod ./
COPY main.go blockchain.go ./

RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -o adapter .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/adapter .

EXPOSE 8090

CMD ["./adapter"]
