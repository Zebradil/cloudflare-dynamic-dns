FROM golang:1.25@sha256:bccbee38ded684484284fdbb72c7d9bc93992f38a0f3b1615a005aa312fa9c4b

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
