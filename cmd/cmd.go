package cmd

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const longDescription = `
Selects an address from the specified network interface and updates A or AAAA
records at Cloudflare for the configured domains. Supports both IPv4 and IPv6.

Required configuration options
--------------------------------------------------------------------------------

--iface:   network interface name to look up for an address
--domains: one or more domain names to assign the address to
--token:   Cloudflare API token with edit access rights to the DNS zone

IPv6 address selection
--------------------------------------------------------------------------------

When multiple IPv6 addresses are found on the interface, the following rules are
used to select the one to use:
    1. Only global unicast addresses (GUA) and unique local addresses (ULA) are
       considered.
    2. GUA addresses are preferred over ULA addresses.
    3. Unique EUI-64 addresses are preferred over randomly generated addresses.
    4. If priority subnets are specified, addresses from the subnet with the
       highest priority are selected. The priority is determined by the order of
       subnets specified on the command line or in the config file.

IPv4 address selection
--------------------------------------------------------------------------------

When multiple IPv4 addresses are found on the interface, the following rules are
used to select the one to use:
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

    iface: eth0
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
`

type stack string

const (
	ipv4 stack = "ipv4"
	ipv6 stack = "ipv6"
)

type runConfig struct {
	domains         []string
	hostId          string
	iface           string
	multihost       bool
	prioritySubnets []string
	proxy           string
	runEvery        time.Duration
	stack           stack
	stateFilepath   string
	token           string
	ttl             int
}

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetEnvPrefix("CFDDNS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if cfgFile == "" {
		cfgFile = viper.GetString("config")
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cloudflare-dynamic-dns")
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
	}
}

func NewRootCmd(version, commit, date string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "cloudflare-dynamic-dns",
		Short:   "Updates A or AAAA records at Cloudflare according to the current address",
		Long:    longDescription,
		Args:    cobra.NoArgs,
		Version: fmt.Sprintf("%s, commit %s, built at %s", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			level, err := log.ParseLevel(viper.GetString("log-level"))
			if err != nil {
				return err
			}
			log.Info(cmd.Name(), " version ", cmd.Version)
			log.Info("Setting log level to:", level)
			log.SetLevel(level)
			return nil
		},
		Run: rootCmdRun,
	}

	rootCmd.
		PersistentFlags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-dynamic-dns.yaml)")

	rootCmd.
		Flags().
		StringSlice("domains", []string{}, "Domain names to assign the address to.")

	rootCmd.
		Flags().
		String("host-id", "", `Unique host identifier. Must be specified in multihost mode.
Must be a valid DNS label. It is stored in the Cloudflare DNS comments field in
the format: "host-id (managed by cloudflare-dynamic-dns)"`)

	rootCmd.
		Flags().
		String("iface", "", "Network interface to look up for an address.")

	rootCmd.
		Flags().
		String("log-level", "info", "Sets logging level: trace, debug, info, warning, error, fatal, panic.")

	rootCmd.
		Flags().
		Bool("multihost", false, `Enable multihost mode.
In this mode it is possible to assign multiple addresses to a single domain.
For correct operation, this mode must be enabled on all participating hosts and
different host-ids must be specified for each host (see --host-id option).`)

	rootCmd.
		Flags().
		StringSlice("priority-subnets", []string{}, `Subnets to prefer over others.
If multiple addresses are found on the interface,
the one from the subnet with the highest priority is used.`)

	rootCmd.
		Flags().
		String("proxy", "auto", `Override proxy setting for created or updated DNS records.
If set to "auto", preserves the current state of an updated record.
Allowed values: "enabled", "disabled", "auto".`)

	rootCmd.
		Flags().
		String("run-every", "", `Re-run the program every N duration until it's killed.
The format is described at https://pkg.go.dev/time#ParseDuration.
The minimum duration is 1m. Examples: 4h30m15s, 5m.`)

	rootCmd.
		Flags().
		String("stack", "ipv6", "IP stack version: ipv4 or ipv6")

	rootCmd.
		Flags().
		String("state-file", "", `Enables usage of a state file.
In this mode, the previously used address is preserved
between runs to avoid unnecessary calls to Cloudflare API.
Automatically selects where to store the state file if no
value is specified. See the State file section in usage.`)

	rootCmd.
		Flags().
		String("token", "", "Cloudflare API token with DNS edit access rights.")

	rootCmd.
		Flags().
		Int("ttl", 1, `Time to live, in seconds, of the DNS record.
Must be between 60 and 86400, or 1 for 'automatic'.`)

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		log.WithError(err).Fatal("Couldn't bind flags")
	}

	return rootCmd
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	cfg := collectConfiguration()
	for {
		run(cfg)
		if cfg.runEvery == 0 {
			break
		}
		log.Info("Sleeping for ", cfg.runEvery)
		time.Sleep(cfg.runEvery)
	}
}

func collectConfiguration() runConfig {
	if viper.ConfigFileUsed() != "" {
		log.WithField("config", viper.ConfigFileUsed()).Debug("Using config file")
		checkConfigAccessMode(viper.ConfigFileUsed())
	} else {
		log.Debug("No config file used")
	}

	var (
		domains          = viper.GetStringSlice("domains")
		hostId           = viper.GetString("host-id")
		iface            = viper.GetString("iface")
		multihost        = viper.GetBool("multihost")
		prioritySubnets  = viper.GetStringSlice("priority-subnets")
		proxy            = viper.GetString("proxy")
		runEvery         = viper.GetString("run-every")
		sleepDuration    = time.Duration(0)
		stack            = stack(viper.GetString("stack"))
		stateFileEnabled = viper.IsSet("state-file")
		stateFilepath    = viper.GetString("state-file")
		token            = viper.GetString("token")
		ttl              = viper.GetInt("ttl")
	)

	if proxy != "auto" && proxy != "enabled" && proxy != "disabled" {
		log.WithField("proxy", proxy).Error("Invalid proxy setting, must be one of: auto, enabled, disabled. Using auto.")
		proxy = "auto"
	}

	if stack != ipv6 && stack != ipv4 {
		log.WithField("stack", stack).Error("Invalid IP mode, must be one of: ipv4, ipv6. Using ipv6.")
		stack = ipv6
	}

	if ttl < 60 || ttl > 86400 {
		// NOTE: 1 is a special value which means "use the default TTL"
		if ttl != 1 {
			log.WithFields(log.Fields{"ttl": ttl}).Warn("TTL must be between 60 and 86400; using Cloudflare's default")
			ttl = 1
		}
	}

	if runEvery != "" {
		parsedDuration, err := time.ParseDuration(runEvery)
		if err != nil {
			log.WithError(err).Fatal("Can't parse provided run-every duration")
		}
		if parsedDuration >= time.Minute {
			sleepDuration = parsedDuration
		} else {
			log.Warn("Provided run-every duration is less then 1 minute, will run just once")
		}
	}

	if stateFileEnabled && stateFilepath == "" {
		domainHash := fnv.New64a()
		domainHash.Write([]byte(strings.Join(domains, " ")))
		stateFilepath = fmt.Sprintf("%s_%s_%x", iface, stack, domainHash.Sum64())
		// If STATE_DIRECTORY is set, use it as the state file directory,
		// otherwise use the current directory.
		if stateDir := os.Getenv("STATE_DIRECTORY"); stateDir != "" {
			stateFilepath = filepath.Join(stateDir, stateFilepath)
		} else {
			log.Info("STATE_DIRECTORY environment is not set, using the current directory for the state file")
		}
	}

	if multihost && hostId != "" {
		hostId = fmt.Sprintf("%s (managed by cloudflare-dynamic-dns)", hostId)
	}

	cfg := runConfig{
		domains:         domains,
		hostId:          hostId,
		iface:           iface,
		multihost:       multihost,
		prioritySubnets: prioritySubnets,
		proxy:           proxy,
		runEvery:        sleepDuration,
		stack:           stack,
		stateFilepath:   stateFilepath,
		token:           token,
		ttl:             ttl,
	}

	logConfig(cfg)

	if token == "" {
		log.Fatal("No token specified")
	}

	if iface == "" {
		log.Fatal("No interface specified")
	}

	if len(domains) == 0 {
		log.Fatal("No domains specified")
	}

	if multihost && hostId == "" {
		log.Fatal("Multihost mode requires host-id to be specified")
	}

	return cfg
}

func logConfig(cfg runConfig) {
	log.WithFields(log.Fields{
		"domains":         cfg.domains,
		"hostId":          cfg.hostId,
		"iface":           cfg.iface,
		"multihost":       cfg.multihost,
		"prioritySubnets": cfg.prioritySubnets,
		"proxy":           cfg.proxy,
		"runEvery":        cfg.runEvery,
		"stack":           cfg.stack,
		"stateFilepath":   cfg.stateFilepath,
		"token":           fmt.Sprintf("[%d characters]", len(cfg.token)),
		"ttl":             cfg.ttl,
	}).Debug("Configuration")
}

func run(cfg runConfig) {
	ipMgr := newIpManager(cfg)
	ip := ipMgr.getIP()

	if cfg.stateFilepath != "" && ip == ipMgr.getOldIp() {
		log.Info("The address hasn't changed, nothing to do")
		log.Info("To bypass this check run without the --state-file flag or remove the state file: ", cfg.stateFilepath)
		return
	}

	api, err := cloudflare.NewWithAPIToken(cfg.token)
	if err != nil {
		log.WithError(err).Fatal("Couldn't create API client")
	}

	for _, domain := range cfg.domains {
		log.Info("Processing domain: ", domain)
		processDomain(api, domain, ip, cfg)
	}

	if cfg.stateFilepath != "" {
		ipMgr.setOldIp(ip)
	}
}

func checkConfigAccessMode(configFilename string) {
	info, err := os.Stat(configFilename)
	if err != nil {
		log.WithError(err).Fatal("Can't get config file info")
	}
	log.WithField("mode", info.Mode()).Debug("Config file mode")
	if info.Mode()&0o011 != 0 {
		log.Warn("Config file should be accessible only by owner")
	}
}

// Ptr returns a pointer to the value passed as an argument.
// This is a workaround for the lack of support for pointer literals in Go.
// See https://stackoverflow.com/a/30716481/2227895 for more information.
func Ptr[T any](v T) *T {
	return &v
}
