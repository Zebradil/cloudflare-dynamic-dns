FROM golang:1.27@sha256:192b74998e350966280a2cbffbb6c40064754f7ec005096aa64f04d7ece4467e

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
