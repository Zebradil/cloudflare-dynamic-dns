FROM golang:1.25@sha256:6ea52a02734dd15e943286b048278da1e04eca196a564578d718c7720433dbbe

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
