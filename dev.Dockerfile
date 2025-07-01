FROM golang:1.24@sha256:a92f3b1ea096cefbe8ec9b51ec11e52f1dc2a728112228411de8970eb3fe7bda

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
