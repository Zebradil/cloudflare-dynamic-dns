FROM alpine:3.23.2@sha256:c93cec902b6a0c6ef3b5ab7c65ea36beada05ec1205664a4131d9e8ea13e405d
RUN apk add --no-cache curl=8.12.1-r0
COPY cloudflare-dynamic-dns /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
