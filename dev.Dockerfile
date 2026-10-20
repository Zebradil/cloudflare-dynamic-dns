FROM golang:1.27@sha256:957861185dbf29d93668ace41e011d091c6e779953d47f0316ef2cc58e28fe2b

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
