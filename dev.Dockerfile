FROM golang:1.24@sha256:14fd8a55e59a560704e5fc44970b301d00d344e45d6b914dda228e09f359a088

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
