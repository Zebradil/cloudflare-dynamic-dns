# Dynamic DNS client for Cloudflare

A small tool for updating IPv6 address at Cloudflare DNS with the currently detected address of the specified network interface.

It is provided with systemd service and timer files for automation.

```
Updates AAAA records at Cloudflare according to the current IPv6 address.

Requires a network interface name for a IPv6 address lookup, domain name[s]
and Cloudflare API token with edit access rights to corresponding DNS zone.

Usage:
  cloudflare-dynamic-dns [flags]

Flags:
      --config string      config file (default is $HOME/.cloudflare-dynamic-dns.yaml)
      --domains strings    Domain names to assign the IPv6 address to
  -h, --help               help for cloudflare-dynamic-dns
      --iface string       Network interface to look up for a IPv6 address
      --log-level string   Sets logging level: trace, debug, info, warning, error, fatal, panic (default "info")
      --systemd            Switch operation mode for running in systemd
                           In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API
      --token string       Cloudflare API token with DNS edit access rights
      --ttl int            Time to live, in seconds, of the DNS record. Must be between 60 and 86400, or 1 for 'automatic' (default 1)
```

## Installation

### AUR

There is a [package in AUR](https://aur.archlinux.org/packages/cloudflare-dynamic-dns/), which can be used on Arch-based distros.

```shell
yay -S cloudflare-dynamic-dns
```

### Manual

Download archive for your OS from [releases page](https://github.com/Zebradil/cloudflare-dynamic-dns/releases).

Or get the source code and build the binary:

```shell
git clone https://github.com/Zebradil/cloudflare-dynamic-dns.git
# OR
curl -sL https://github.com/Zebradil/cloudflare-dynamic-dns/archive/refs/heads/master.tar.gz | tar xz

cd cloudflare-dynamic-dns-master
go build -o cloudflare-dynamic-dns main.go
```

Now you can run `cloudflare-dynamic-dns` manually (see [Usage](#usage) section).

If you want to do some automation with systemd, `cloudflare-dynamic-dns` has to be installed system-wide
(it *is* possible to run systemd timer and service without root privileges, but I do not provide ready-to-use configuration for this yet):

```shell
sudo install -Dm755 cloudflare-dynamic-dns -t /usr/bin 
sudo install -Dm644 systemd/* -t /usr/lib/systemd/system
sudo install -m700 -d /etc/cloudflare-dynamic-dns/config.d
```

## Usage

### Run manually

0. Follow steps from [Installation](#instllation) section
1. Run `./cloudflare-dynamic-dns --domains 'example.com,*.example.com' --iface eth0 --token cloudflare-api-token`
   - NOTE: instead of compiling `cloudflare-dynamic-dns` binary, it can be replaced with `go run main.go` in the command above.

Instead of specifying command line arguments, it is possible to create `~/.cloudflare-dynamic-dns.yaml` with the following structure:

```yaml
iface: eth0
token: cloudflare-api-token
domains:
  - example.com
  - "*.example.com"
```

And then run `./cloudflare-dynamic-dns` (or `go run main.go`) without arguments.
Or put the configuration in any place and specify it with `--config` flag (that's not tested):

```shell
./cloudflare-dynamic-dns --config /any/place/config.yaml
```

### Systemd service and timer

It is possible to run `cloudflare-dynamic-dns` periodically via systemd.
This requires privileged access to the system.
Make sure that required systemd files are installed (see [Installation](#instllation) section for details).

```shell
# 1. Create configuration file `/etc/cloudflare-dynamic-dns/config.d/<name>.yaml`
#    For example (I use "example.com" as <name>, replace the values according to your needs):
sudo tee -a /etc/cloudflare-dynamic-dns/config.d/example.com.yaml <<EOF
iface: eth0
token: cloudflare-api-token
domains:
  - example.com
  - "*.example.com"
EOF

# 3. Enable systemd timer
sudo systemctl enable --now cloudflare-dynamic-dns@example.com.timer
```

This way (via running multiple timers) you can use multiple configurations at the same time.

By default a timer is triggered one minute after boot and then every 5 minutes. It is not configurable currently.

To avoid unnecessary requests to Cloudflare API state files are used.
They're created in `/var/lib/cloudflare-dynamic-dns/` and named using configuration variable in corresponding config files (`iface` and md5 hash of `domains`).
A state file contains IPv6 address which was set in a Cloudflare DNS AAAA record during the last successful run.
If the current IPv6 address is the same as in the one in the state file, no additional API requests are done.

## Development

Builds and releases are done with [goreleaser](https://goreleaser.com/).

Use the following Taskfile tasks to bulid the application:

```shell
# Build with go for the current platform
task build

# Build with GoReleaser for all configured platforms
task goreleaser:build
```

## License

MIT
