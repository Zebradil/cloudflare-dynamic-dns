FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
RUN apk add --no-cache curl=8.12.1-r0
COPY cloudflare-dynamic-dns /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
