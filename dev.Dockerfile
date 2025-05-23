FROM golang:1.24@sha256:795a40cbe36a11e39b0709eb98d7aaa8d312d60336d863a0ef1c8aff07c1b3e0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
