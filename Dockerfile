FROM scratch
COPY --from=alpine:20240807 --link /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY cloudflare-dynamic-dns /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
