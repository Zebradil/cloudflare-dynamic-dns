FROM golang:1.27@sha256:921f61120011c774a7b37fd80bd3c3d259636090d3cd002b6a344322a0328082

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
