FROM golang:1.25@sha256:91e2cd436f7adbfad0a0cbb7bf8502fa863ed8461414ceebe36c6304731e0fd9

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
