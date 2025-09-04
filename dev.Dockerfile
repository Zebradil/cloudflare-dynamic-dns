FROM golang:1.25@sha256:5e856b892fff06368bd92fe3ac00c457ede6d399e7152575b57f9a14312ab713

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
