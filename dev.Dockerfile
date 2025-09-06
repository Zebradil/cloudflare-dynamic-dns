FROM golang:1.25@sha256:a5e935dbd8bc3a5ea24388e376388c9a69b40628b6788a81658a801abbec8f2e

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
