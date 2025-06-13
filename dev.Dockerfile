FROM golang:1.24@sha256:10c131810f80a4802c49cab0961bbe18a16f4bb2fb99ef16deaa23e4246fc817

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
