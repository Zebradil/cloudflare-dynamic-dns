FROM alpine:3.24.4@sha256:93b41ef40b1677fce8fcc64ff4af09daddfacf5040afde618bc46c1e2d1cc7d4
RUN apk add --no-cache curl
COPY $TARGETPLATFORM/cloudflare-dynamic-dns /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
