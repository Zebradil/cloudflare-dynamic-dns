FROM golang:1.24@sha256:8131d99bd9bdf38a3b9720c23c3d21f40d65137d4b2fd2eccab20e7ab7b5dee4

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
