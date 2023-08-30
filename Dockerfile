FROM scratch
COPY . /
ENTRYPOINT ["/cloudflare-dynamic-dns"]
