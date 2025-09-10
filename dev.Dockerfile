FROM golang:1.25@sha256:1fd7d46f956287d1856b92add5cc5ab8b87c07a1ed766419bb603a8620746b4a

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
