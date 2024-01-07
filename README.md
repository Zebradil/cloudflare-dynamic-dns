# Dynamic DNS client for Cloudflare

A small tool for updating the IPv6 address at Cloudflare DNS with the currently detected address of the specified network interface.

## Features

- Intellegent IPv6 address selection
- Support for multiple domains
- Systemd service and timer files for automation
- Can be run as a Docker containern
- Configuration via command line arguments, config file or environment variables
- Support for multiple hosts in the same domain (experimental)

```text
Selects an IPv6 address from the specified network interface and updates AAAA
records at Cloudflare for the configured domains.

=== Required configuration options ===

--iface:   network interface name to look up for an IPv6 address
--domains: one or more domain names to assign the IPv6 address to
--token:   Cloudflare API token with edit access rights to the DNS zone

=== IPv6 address selection ===

When multiple IPv6 addresses are found on the interface, the following rules are
used to select the one to use:
    1. Only global unicast addresses (GUA) and unique local addresses (ULA) are
       considered.
    2. GUA addresses are preferred over ULA addresses.
    3. Unique EUI-64 addresses are preferred over randomly generated addresses.
    4. If priority subnets are specified, addresses from the subnet with the
       highest priority are selected. The priority is determined by the order of
       subnets specified on the command line or in the config file.

=== Daemon/systemd mode ===

The program can be run in systemd mode, in which case the previously used IPv6
address is preserved between runs to avoid unnecessary calls to the Cloudflare
API. This mode is enabled by passing --systemd flag. The state file is stored
in the directory specified by the STATE_DIRECTORY environment variable.

=== Multihost mode (EXPERIMENTAL) ===

In this mode it is possible to assign multiple IPv6 addresses to a single or
multiple domains. For correct operation, this mode must be enabled on all hosts
participating in the same domain and different host-ids must be specified for
each host (see --host-id option). This mode is enabled by passing --multihost
flag.

In the multihost mode, the program will manage only the DNS records that have
the same host-id as the one specified on the command line or in the config file.
Any other records will be ignored. This allows multiple hosts to share the same
domain without interfering with each other. The host-id is stored in the
Cloudflare DNS comments field (see https://developers.cloudflare.com/dns/manage-dns-records/reference/record-attributes/).

=== Persistent configuration ===

The program can be configured using a config file. The default location is
$HOME/.cloudflare-dynamic-dns.yaml. The config file location can be overridden
using the --config flag. The config file format is YAML. The following options
are supported (with example values):

    iface: eth0
    token: cloudflare-api-token
    domains:
      - example.com
      - "*.example.com"
    # --- optional fields ---
    logLevel: info
    prioritySubnets:
      - 2001:db8::/32
      - 2001:db8:1::/48
    ttl: 180
    systemd: false
    multihost: true
    hostId: homelab-node-1

=== Environment variables ===

The configuration options can be specified as environment variables. To make an
environment variable name, prefix a flag name with CFDDNS_, replace dashes with
underscores, and convert to uppercase. List values are specified as a single
string containing elements separated by spaces.
For example:

    CFDDNS_IFACE=eth0
    CFDDNS_TOKEN=cloudflare-api-token
    CFDDNS_DOMAINS='example.com *.example.com'
    CFDDNS_LOG_LEVEL=info
    CFDDNS_PRIORITY_SUBNETS='2001:db8::/32 2001:db8:1::/48'
    CFDDNS_TTL=180
    CFDDNS_SYSTEMD=false
    CFDDNS_MULTIHOST=true
    CFDDNS_HOST_ID=homelab-node-1

Usage:
  cloudflare-dynamic-dns [flags]

Flags:
      --config string              config file (default is $HOME/.cloudflare-dynamic-dns.yaml)
      --domains strings            Domain names to assign the IPv6 address to.
  -h, --help                       help for cloudflare-dynamic-dns
      --host-id string             Unique host identifier. Must be specified in multihost mode.
      --iface string               Network interface to look up for a IPv6 address.
      --log-level string           Sets logging level: trace, debug, info, warning, error, fatal, panic. (default "info")
      --multihost                  Enable multihost mode.
                                   In this mode it is possible to assign multiple IPv6 addresses to a single domain.
                                   For correct operation, this mode must be enabled on all participating hosts and
                                   different host-ids must be specified for each host (see --host-id option).
      --priority-subnets strings   IPv6 subnets to prefer over others.
                                   If multiple IPv6 addresses are found on the interface, the one from the subnet with the highest priority is used.
      --systemd                    Switch operation mode for running in systemd.
                                   In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API.
      --token string               Cloudflare API token with DNS edit access rights.
      --ttl int                    Time to live, in seconds, of the DNS record. Must be between 60 and 86400, or 1 for 'automatic'. (default 1)
  -v, --version                    version for cloudflare-dynamic-dns
```

## Installation

### AUR

There are two packages in AUR (
[1](https://aur.archlinux.org/packages/cloudflare-dynamic-dns/),
[2](https://aur.archlinux.org/packages/cloudflare-dynamic-dns-bin/)
), that can be used on Arch-based distros:

```shell
yay -S cloudflare-dynamic-dns
# OR
yay -S cloudflare-dynamic-dns-bin
```

### Docker

See the
[container registry page](https://github.com/Zebradil/cloudflare-dynamic-dns/pkgs/container/cloudflare-dynamic-dns)
for details.

```shell
docker pull ghcr.io/zebradil/cloudflare-dynamic-dns:latest
```

### DEB, RPM, APK

See the [latest release page](https://github.com/Zebradil/cloudflare-dynamic-dns/releases/latest) for the full list of
packages.

### Manual

Download the archive for your OS from the [releases page](https://github.com/Zebradil/cloudflare-dynamic-dns/releases).

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
(it _is_ possible to run systemd timer and service without root privileges, but I do not provide ready-to-use configuration for this yet):

```shell
sudo install -Dm755 cloudflare-dynamic-dns -t /usr/bin
sudo install -Dm644 systemd/* -t /usr/lib/systemd/system
sudo install -m700 -d /etc/cloudflare-dynamic-dns/config.d
```

## Usage

### Run manually

0. Follow the steps from the [Installation](#installation) section.
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
Or put the configuration in any place and specify it with `--config` flag:

```shell
./cloudflare-dynamic-dns --config /any/place/config.yaml
```

### Priority subnets

If multiple IPv6 addresses are found on the interface, the one from the subnet with the highest priority is used.
If no priority subnets are specified, the first address is used.
Priority subnets are specified in the configuration file:

```yaml
prioritySubnets:
  - 2001:db8::/32
  - 2001:db8:1::/48
```

Or via `--priority-subnets` flag:

```shell
./cloudflare-dynamic-dns --priority-subnets '2001:db8::/32,2001:db8:1::/48'
```

### Systemd service and timer

It is possible to run `cloudflare-dynamic-dns` periodically via systemd.
This requires privileged access to the system.
Make sure that required systemd files are installed (see [Installation](#installation) section for details).

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

By default, a timer is triggered one minute after boot and then every 5 minutes. It is not configurable currently.

To avoid unnecessary requests to Cloudflare API state files are used.
They're created in `/var/lib/cloudflare-dynamic-dns/` and named using configuration variables in corresponding config files (`iface` and md5 hash of `domains`).
A state file contains IPv6 address which was set in a Cloudflare DNS AAAA record during the last successful run.
If the current IPv6 address is the same as the one in the state file, no additional API requests are done.

## Development

Builds and releases are done with [goreleaser](https://goreleaser.com/).

Use the following Taskfile tasks to build the application:

```shell
# Build with go for the current platform
go build -o cloudflare-dynamic-dns main.go

# Build with GoReleaser for all configured platforms
task build

# Use Docker
docker build -t cloudflare-dynamic-dns -f dev.Dockerfile .
```

### GeReleaser

Do not change `.goreleaser.yml` manually, do changes in `.goreleaser.ytt.yml` and run
`task misc:build:goreleaser-config` instead (requires [`ytt`](https://carvel.dev/ytt/) installed).

## License

[MIT](LICENSE)

## Disclaimer

This project is not affiliated with Cloudflare.
