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
      --systemd         Switch operation mode for running in systemd
                        In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API
      --token string    Cloudflare API token with DNS edit access rights
```

## Usage

### Run manually

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

### Systemd service and timer (not tested yet)

It is possible to run `cloudflare-dynamic-dns` periodically via systemd. Requires privileged access to the system.

```shell
# 1. Copy `systemd/cloudflare-dynamic-dns.service` and `systemd/cloudflare-dynamic-dns.timer` to `/usr/lib/systemd/system`
sudo cp systemd/* /usr/lib/systemd/system/

# 2. Create configuration file `/etc/cloudflare-dynamic-dns/config.d/<name>.yaml`
#    For exaple (replace the values according to your needs):
sudo tee -a /etc/cloudflare-dynamic-dns/config.d/example.com.yaml <<EOF
iface: eth0
token: cloudflare-api-token
domain: example.com
EOF

# 3. Enable systemd timer
sudo systemd enable --now cloudflare-dynamic-dns@example.com.timer
```

This way (via running multiple timers) you can use multiple configurations at the same time.

By default a timer is triggered one minute after boot and then every 5 minutes. It is not configurable currently.
