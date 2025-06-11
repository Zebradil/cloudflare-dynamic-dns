FROM golang:1.24@sha256:01f861b85f25cd045191578c87ed057b92bd185f2226bbf375c7510ebcb8abf4

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
