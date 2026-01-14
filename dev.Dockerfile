FROM golang:1.25@sha256:581c059c1d53b96a6db51a2f7a0eb943491ba1da50f1e43700db0cc325618096

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
