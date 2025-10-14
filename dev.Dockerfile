FROM golang:1.25@sha256:c0bf2bc2f8e5720aa2e83e45d2085edbf2ad085e2d1a195bb6c3c402350fe661

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
