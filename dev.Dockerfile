FROM golang:1.25@sha256:d2e5acc5c29cc331ad5e4f59b09dee0f7d47043b072de0799be63e5c49f95feb

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
