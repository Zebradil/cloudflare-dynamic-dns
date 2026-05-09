FROM golang:1.26@sha256:13605dbaf3aff39741644cc3a6ec74cac494f955d1401ffee49b55032fa8a626

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
