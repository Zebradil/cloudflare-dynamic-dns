FROM golang:1.24@sha256:884849e632f7b90be1acd3293579dbf19595c582a202b98411c88fdb60a319f0

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
