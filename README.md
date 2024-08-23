# Dynamic DNS client for Cloudflare

A CLI tool for updating A/AAAA record at Cloudflare DNS with the currently detected address of the specified network interface.

## Features

- Supports:
  - IPv4 and IPv6
  - Multiple domains with the same address
  - Multiple hosts in the same domain
- Tries to be smart about selecting the address to use
- Includes systemd service and timer files for automation
- Can be run in a Docker container
- Configuration via command line arguments, config file or environment variables

## Usage

The rest of this section is the output of `cloudflare-dynamic-dns --help`.

<!-- BEGIN CFDDNS_USAGE -->
<pre>

Selects an address from the specified network interface or via an external
command and updates A or AAAA records at Cloudflare for the configured domains.
Supports both IPv4 and IPv6.

Required configuration options
--------------------------------------------------------------------------------

--iface:   network interface name to look up for an address
  or
--ipcmd:   shell command to run to get the address, should return one address
           per line. Uses https://github.com/mvdan/sh as the shell.
           Examples:
             - curl -fsSL https://api6.ipify.org
             - echo -e "127.0.0.1\n127.0.0.2"

--domains: one or more domain names to assign the address to
--token:   Cloudflare API token with edit access rights to the DNS zone

IPv6 address selection
--------------------------------------------------------------------------------

When multiple IPv6 addresses are found on the interface or received from the
external command (e.g., when using --ipcmd), the following rules are used to
select the one to use:
    1. Only global unicast addresses (GUA) and unique local addresses (ULA) are
       considered.
    2. GUA addresses are preferred over ULA addresses.
    3. Unique EUI-64 addresses are preferred over randomly generated addresses.
    4. If priority subnets are specified, addresses from the subnet with the
       highest priority are selected. The priority is determined by the order of
       subnets specified on the command line or in the config file.

IPv4 address selection
--------------------------------------------------------------------------------

When multiple IPv4 addresses are found on the interface or received from the
external command (e.g., when using --ipcmd), the following rules are used to
select the one to use:
    1. All IPv4 addresses are considered.
    2. Public addresses are preferred over Shared Address Space (RFC 6598)
       addresses.
    3. Shared Address Space addresses are preferred over private addresses.
    4. Private addresses are preferred over loopback addresses.
    5. If priority subnets are specified, addresses from the subnet with the
       highest priority are selected. The priority is determined by the order of
       subnets specified on the command line or in the config file.

Non-public addresses are logged as warnings but are still used. They can be
useful in private networks or when using a VPN.

NOTE: Cloudflare doesn't allow proxying of records with non-public addresses.

Daemon mode
--------------------------------------------------------------------------------

By default, the program runs once and exits. This mode of operation can be
changed by setting the --run-every flag to a duration greater than 1m. In this
case, the program will run repeatedly, waiting the duration between runs. It
will stop if killed or if failed.

State file
--------------------------------------------------------------------------------

Setting --state-file makes the program to retain the previously used address
between runs to avoid unnecessary calls to the Cloudflare API.

The value is used as the state file path. When used with an empty value, the
state file is named after the interface name and the domains, and is stored
either in the current directory or in the directory specified by the
STATE_DIRECTORY environment variable.

The STATE_DIRECTORY environment variable is automatically set by systemd. It
can be set manually when running the program outside of systemd.

Multihost mode (EXPERIMENTAL)
--------------------------------------------------------------------------------

In this mode, it is possible to assign multiple addresses to a single or
multiple domains. For correct operation, this mode must be enabled on all hosts
participating in the same domain and different host-ids must be specified for
each host (see --host-id option). This mode is enabled by passing --multihost
flag.

In the multihost mode, the program will manage only the DNS records that have
the same host-id as the one specified on the command line or in the config file.
If an existing record has no host-id but has the same address as the target one,
it will be claimed by this host via setting the corresponding host-id. Any other
records will be ignored. This allows multiple hosts to share the same domain
without interfering with each other. The host-id is stored in the Cloudflare DNS
comments field (see https://developers.cloudflare.com/dns/manage-dns-records/reference/record-attributes/).

Persistent configuration
--------------------------------------------------------------------------------

The program can be configured using a config file. The default location is
$HOME/.cloudflare-dynamic-dns.yaml. The config file location can be overridden
using the --config flag. The config file format is YAML. The following options
are supported (with example values):

    # === required fields
    # either iface or ipcmd must be specified
    iface: eth0
    # ipcmd: curl -fsSL https://api6.ipify.org
    token: cloudflare-api-token
    domains:
      - example.com
      - "*.example.com"
    # === optional fields
    # --- mode
    stack: ipv6
    # --- UI
    log-level: info
    # --- logic
    priority-subnets:
      - 2001:db8::/32
      - 2001:db8:1::/48
    multihost: true
    host-id: homelab-node-1
    # --- DNS record details
    proxy: enabled
    ttl: 180
    # --- daemon mode
    run-every: 10m
    state-file: /tmp/cfddns-eth0.state

Environment variables
--------------------------------------------------------------------------------

The configuration options can be specified as environment variables. To make an
environment variable name, prefix a flag name with CFDDNS_, replace dashes with
underscores, and convert to uppercase. List values are specified as a single
string containing elements separated by spaces.
For example:

    CFDDNS_CONFIG=/path/to/config.yaml
    CFDDNS_IFACE=eth0
    CFDDNS_IPCMD='curl -fsSL https://api6.ipify.org'
    CFDDNS_TOKEN=cloudflare-api-token
    CFDDNS_DOMAINS='example.com *.example.com'
    CFDDNS_STACK=ipv6
    CFDDNS_LOG_LEVEL=info
    CFDDNS_PRIORITY_SUBNETS='2001:db8::/32 2001:db8:1::/48'
    CFDDNS_MULTIHOST=true
    CFDDNS_HOST_ID=homelab-node-1
    CFDDNS_PROXY=enabled
    CFDDNS_TTL=180
    CFDDNS_RUN_EVERY=10m
    CFDDNS_STATE_FILE=/tmp/cfddns-eth0.state

Usage:
  cloudflare-dynamic-dns [flags]

Flags:
      --config string              config file (default is $HOME/.cloudflare-dynamic-dns.yaml)
      --domains strings            Domain names to assign the address to.
  -h, --help                       help for cloudflare-dynamic-dns
      --host-id string             Unique host identifier. Must be specified in multihost mode.
                                   Must be a valid DNS label. It is stored in the Cloudflare DNS comments field in
                                   the format: "host-id (managed by cloudflare-dynamic-dns)"
      --iface string               Network interface to look up for an address.
      --ipcmd string               External command to run to get the address.
      --log-level string           Sets logging level: trace, debug, info, warning, error, fatal, panic. (default "info")
      --multihost                  Enable multihost mode.
                                   In this mode it is possible to assign multiple addresses to a single domain.
                                   For correct operation, this mode must be enabled on all participating hosts and
                                   different host-ids must be specified for each host (see --host-id option).
      --priority-subnets strings   Subnets to prefer over others.
                                   If multiple addresses are found on the interface,
                                   the one from the subnet with the highest priority is used.
      --proxy string               Override proxy setting for created or updated DNS records.
                                   If set to "auto", preserves the current state of an updated record.
                                   Allowed values: "enabled", "disabled", "auto". (default "auto")
      --run-every string           Re-run the program every N duration until it's killed.
                                   The format is described at https://pkg.go.dev/time#ParseDuration.
                                   The minimum duration is 1m. Examples: 4h30m15s, 5m.
      --stack string               IP stack version: ipv4 or ipv6 (default "ipv6")
      --state-file string          Enables usage of a state file.
                                   In this mode, the previously used address is preserved
                                   between runs to avoid unnecessary calls to Cloudflare API.
                                   Automatically selects where to store the state file if no
                                   value is specified. See the State file section in usage.
      --token string               Cloudflare API token with DNS edit access rights.
      --ttl int                    Time to live, in seconds, of the DNS record.
                                   Must be between 60 and 86400, or 1 for 'automatic'. (default 1)
  -v, --version                    version for cloudflare-dynamic-dns
</pre>
<!-- END CFDDNS_USAGE -->

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

## Usage examples

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
  - '*.example.com'
```

And then run `./cloudflare-dynamic-dns` (or `go run main.go`) without arguments.
Or put the configuration in any place and specify it with `--config` flag:

```shell
./cloudflare-dynamic-dns --config /any/place/config.yaml
```

#### Run in a Docker container

For the binary, the usage is the same as for the manual run.
For the Docker container, we need to mount the configuration file into the container and provide access to the network stack of the host machine:

```shell
docker run --rm \
  --volume="/any/place/config.yaml:/config.yaml" \
  --network=host \
  ghcr.io/zebradil/cloudflare-dynamic-dns:latest \
    --config=/config.yaml
```

To run the program in daemon mode, add `--run-every` flag (and `--detach` if you want to run it in the background):

```shell
docker run --rm \
  --volume="/any/place/config.yaml:/config.yaml" \
  --network=host \
  --detach \
  ghcr.io/zebradil/cloudflare-dynamic-dns:latest \
    --config=/config.yaml \
    --run-every=5m
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

State files are used to avoid unnecessary requests to Cloudflare API.
They're created in `/var/lib/cloudflare-dynamic-dns/` and named using configuration variables in corresponding config files (`iface` and a hash of `domains`).
A state file contains an address which was set in a Cloudflare DNS AAAA or A record during the last successful run.
If the current address is the same as the one in the state file, no additional API requests are done.

## Development

Builds and releases are done with [goreleaser](https://goreleaser.com/).

There are several ways to build the application:

```shell
# Build with go for the current platform
go build -o cloudflare-dynamic-dns main.go

# Build with GoReleaser for all configured platforms
task go:build

# Use Docker
docker build -t cloudflare-dynamic-dns -f dev.Dockerfile .
```

Check the [Taskfile.yml](./Taskfile.yml) for more details.

### GeReleaser

:warning: Do not change `.goreleaser.yml` manually, do changes in `.goreleaser.ytt.yml` and run
`task misc:build:goreleaser-config` instead (requires [`ytt`](https://carvel.dev/ytt/) installed).

### Documentation

The [usage](#usage) section is generated by the [update_readme](./scripts/update_readme) script.
For convenience, `task docs:update-readme` can be used to run it.

### Vendor dependencies

Most of the dependencies are managed with Go modules, but there is one exception: the [execext](internal/execext)
package, which is an internal package of the [go-task](https://github.com/go-task/task) project and is vendored using
[vendir](https://github.com/carvel-dev/vendir).

To update the vendored dependencies, run `vendir sync` in the root of the project. Then commit `vendir.lock.yml` and the
updated dependencies.

## License

[MIT](LICENSE)

Code of the [execext](internal/execext) package is taken from the [go-task](https://github.com/go-task/task) project and
is licensed under the [MIT license](https://github.com/go-task/task/blob/0941de3318ba2cd8e8611acc4e225b0bea5c817b/LICENSE).

## Disclaimer

This project is not affiliated with Cloudflare.
