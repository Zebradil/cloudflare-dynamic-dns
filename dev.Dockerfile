FROM golang:1.25@sha256:ce63a16e0f7063787ebb4eb28e72d477b00b4726f79874b3205a965ffd797ab2

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
