FROM golang:1.25@sha256:0caf875670e0ec9ebe7f4a9f4cf02add9d06ffccb055cc1066c83270c237dfb9

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
