FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o netsim .

FROM scratch
WORKDIR /app
COPY --from=builder /app/netsim .
COPY --from=builder /app/config.json .
EXPOSE 3000/udp
ENTRYPOINT ["./netsim"]
