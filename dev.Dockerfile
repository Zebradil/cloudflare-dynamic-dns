FROM golang:1.25@sha256:ab1f5c47de0f2693ed97c46a646bde2e4f380e40c173454d00352940a379af60

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
