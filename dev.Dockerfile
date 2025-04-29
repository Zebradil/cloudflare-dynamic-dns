FROM golang:1.24@sha256:3a060d683c28fbb21d7fe8966458e084a6d7ebfb1f3ef3fd901abd2083c43675

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
