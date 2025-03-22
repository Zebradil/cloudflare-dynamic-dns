FROM golang:1.24@sha256:52ff1b35ff8de185bf9fd26c70077190cd0bed1e9f16a2d498ce907e5c421268

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
