FROM scratch
COPY --from=alpine:20230901 --link /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY . /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
