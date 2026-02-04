FROM golang:1.25@sha256:06d1251c59a75761ce4ebc8b299030576233d7437c886a68b43464bad62d4bb1

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
