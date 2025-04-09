FROM golang:1.24@sha256:c0b66cfec8562c8d8452c63a63ede2014526224fa1cd5ecf5acda73f8b28263b

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
