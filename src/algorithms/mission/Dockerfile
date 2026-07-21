FROM golang:1.26.1-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .

# Execution arguments (like config files) must be passed dynamically when running the container
ENTRYPOINT ["./main", "config.json"]
