FROM golang:1.26@sha256:6df14f4a4bc9d979a3721f488981e0d1b318006377e473ed23d026796f5f4c0a

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
