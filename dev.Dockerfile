FROM golang:1.25@sha256:516827db2015144cf91e042d1b6a3aca574d013a4705a6fdc4330444d47169d5

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
