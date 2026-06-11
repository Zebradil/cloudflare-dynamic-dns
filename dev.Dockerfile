FROM golang:1.26@sha256:d47ca13cd596f3a338c1be5f79af628f42bedcf89455266211a9ab4f95da2828

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
