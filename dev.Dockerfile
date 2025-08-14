FROM golang:1.24@sha256:61808652990bcaa6981db6a85ecd0099c8fa10a6d49c3bd40194c00b69917856

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
