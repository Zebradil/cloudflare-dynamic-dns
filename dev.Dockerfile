FROM golang:1.26@sha256:98fc714bfe32e7d3c539d63bda9b9cd089fd699dc3cbd1c534fec3c4deb9ca98

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
