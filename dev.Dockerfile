FROM golang:1.24@sha256:034848561f95a942e2163d9017e672f0c65403f699336db4529a908af00dfc98

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
