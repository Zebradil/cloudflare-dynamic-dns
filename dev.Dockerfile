FROM golang:1.27@sha256:6f1acc898d064a4e9d5773ee574a9af6a9b786306939ac14d7fd806bf62046db

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
