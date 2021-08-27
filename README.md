# Dynamic DNS client for Cloudflare

```
Updates AAAA records at Cloudflare according to the current IPv6 address.

Requires a network interface name for a IPv6 address lookup, domain name
and Cloudflare API token with edit access rights to corresponding DNS zone.

Usage:
  cloudflare-dynamic-dns [flags]

Flags:
      --config string   config file (default is $HOME/.cloudflare-dynamic-dns.yaml)
      --domain string   Domain name to assign the IPv6 address to
  -h, --help            help for cloudflare-dynamic-dns
      --iface string    Network interface to look up for a IPv6 address
      --token string    Cloudflare API token with DNS edit access rights
```

## Usage

1. Clone the repo (releases and packages are WIP)
2. Run `go run main.go --domain example.com --iface eth0 --token cloudflare-api-token`

Instead of specifying command line arguments, it is possible to create `~/.cloudflare-dynamic-dns.yaml` with the following structure:

```yaml
iface: eth0
token: cloudflare-api-token
domain: example.com
```

And then run `go run main.go` (without arguments).
Or put the configuration in any place and specify it with `--config` flag (that's not tested):

```
go run main.go --config /any/place/config.yaml
```
