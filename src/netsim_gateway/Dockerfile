FROM golang:1.26.2-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o netsim_gateway .

FROM scratch
WORKDIR /app
COPY --from=builder /app/netsim_gateway .
COPY --from=builder /app/config.json .
EXPOSE 3000/udp
EXPOSE 3001/udp
ENTRYPOINT ["./netsim_gateway"]
