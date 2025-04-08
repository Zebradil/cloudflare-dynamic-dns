FROM golang:1.24@sha256:fb224f950b0dfb889f8f1122bf3dfc21976735ccf983b73a1e6e014215a272f5

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
