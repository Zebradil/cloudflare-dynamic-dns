FROM golang:1.25@sha256:1c91b4f4391774a73d6489576878ad3ff3161ebc8c78466ec26e83474855bfcf

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
