FROM golang:1.26@sha256:11fd8f7f63db3b6fb198797042ba4c40a4a34dc83325d3328ca3bc4bb7726786

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
