FROM alpine:3.21.3
RUN apk add --no-cache curl=8.12.1-r0
COPY cloudflare-dynamic-dns /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
