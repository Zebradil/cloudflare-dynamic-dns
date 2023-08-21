FROM golang:1.19-alpine

WORKDIR /app
COPY . /app

RUN go mod download
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["bin/cloudflare-dynamic-dns"]
