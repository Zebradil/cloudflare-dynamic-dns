FROM golang:1.24@sha256:764d7e0ce1df1e4a1bddc6d1def5f3516fdc045c5fad88e61f67fdbd1857282f

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN go build -o bin/cloudflare-dynamic-dns main.go

ENTRYPOINT ["/app/bin/cloudflare-dynamic-dns"]
