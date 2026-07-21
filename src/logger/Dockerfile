FROM golang:1.26.2-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /logger_bin ./cmd/logger

FROM alpine:latest
WORKDIR /app
COPY --from=builder /logger_bin /app/logger_bin
COPY config.json /app/config.json

EXPOSE 5000/udp 8080/tcp

CMD ["/app/logger_bin", "config.json"]
