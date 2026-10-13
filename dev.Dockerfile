FROM golang:1.27@sha256:007a316059246276cb214df93950be77837545dc78007a73e0291a4963d5d3bf

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
