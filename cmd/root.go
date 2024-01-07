package cmd

import (
	"context"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const longDescription = `
Selects an IPv6 address from the specified network interface and updates AAAA
records at Cloudflare for the configured domains.

Required configuration options
--------------------------------------------------------------------------------

--iface:   network interface name to look up for an IPv6 address
--domains: one or more domain names to assign the IPv6 address to
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

Daemon mode
--------------------------------------------------------------------------------

By default, the program runs once and exits. This mode of operation can be
changed by setting the --run-every flag to a duration greater than 1m. In this
case, the program will run repeatedly, waiting the duration between runs. It
will stop if killed or if failed.

State file
--------------------------------------------------------------------------------

Setting --with-state-file makes the program to retain the previously used IPv6
address between runs to avoid unnecessary calls to the Cloudflare API.

The flag can be used with or without a value. The value is used as the state
file path. When used without a value, the state file is named after the
interface name and the domains, and is stored either in the current directory or
in the directory specified by the STATE_DIRECTORY environment variable.

The STATE_DIRECTORY environment variable is automatically set by systemd. It
can be set manually when running the program outside of systemd.

Multihost mode (EXPERIMENTAL)
--------------------------------------------------------------------------------

In this mode, it is possible to assign multiple IPv6 addresses to a single or
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
    # --- optional fields ---
    logLevel: info
    prioritySubnets:
      - 2001:db8::/32
      - 2001:db8:1::/48
    ttl: 180
    runEvery: 10m
    withStateFile: /tmp/cfddns-eth0.state
    multihost: true
    hostId: homelab-node-1

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
    CFDDNS_LOG_LEVEL=info
    CFDDNS_PRIORITY_SUBNETS='2001:db8::/32 2001:db8:1::/48'
    CFDDNS_TTL=180
    CFDDNS_RUN_EVERY=10m
    CFDDNS_WITH_STATE_FILE=/tmp/cfddns-eth0.state
    CFDDNS_MULTIHOST=true
    CFDDNS_HOST_ID=homelab-node-1
`

type runConfig struct {
	domains         []string
	hostId          string
	iface           string
	multihost       bool
	prioritySubnets []string
	runEvery        time.Duration
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
		Short:   "Updates AAAA records at Cloudflare according to the current IPv6 address",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-dynamic-dns.yaml)")

	rootCmd.Flags().String("with-state-file", "", `Enables usage of a state file.
In this mode, previously used ipv6 address is preserved
between runs to avoid unnecessary calls to CloudFlare API.
Automatically selects where to store the state file if no
value is specified. See the State file section in usage.`)
	rootCmd.Flags().String("run-every", "", `Re-run the program every N duration until it's killed.
The format is described at https://pkg.go.dev/time#ParseDuration.
The minimum duration is 1m. Examples: 4h30m15s, 5m.`)
	rootCmd.Flags().Bool("multihost", false, `Enable multihost mode.
In this mode it is possible to assign multiple IPv6 addresses to a single domain.
For correct operation, this mode must be enabled on all participating hosts and
different host-ids must be specified for each host (see --host-id option).`)
	rootCmd.Flags().String("host-id", "", `Unique host identifier. Must be specified in multihost mode.
Must be a valid DNS label. It is stored in the Cloudflare DNS comments field in
the format: "host-id (managed by cloudflare-dynamic-dns)"`)
	rootCmd.Flags().Int("ttl", 1, `Time to live, in seconds, of the DNS record.
Must be between 60 and 86400, or 1 for 'automatic'.`)
	rootCmd.Flags().StringSlice("domains", []string{}, "Domain names to assign the IPv6 address to.")
	rootCmd.Flags().StringSlice("priority-subnets", []string{}, `IPv6 subnets to prefer over others.
If multiple IPv6 addresses are found on the interface,
the one from the subnet with the highest priority is used.`)
	rootCmd.Flags().String("iface", "", "Network interface to look up for a IPv6 address.")
	rootCmd.Flags().String("log-level", "info", "Sets logging level: trace, debug, info, warning, error, fatal, panic.")
	rootCmd.Flags().String("token", "", "Cloudflare API token with DNS edit access rights.")

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
		prioritySubnets  = viper.GetStringSlice("prioritySubnets")
		runEvery         = viper.GetString("run-every")
		sleepDuration    = time.Duration(0)
		stateFilepath    = viper.GetString("with-state-file")
		stateFileEnabled = viper.IsSet("with-state-file")
		token            = viper.GetString("token")
		ttl              = viper.GetInt("ttl")
	)

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
		stateFilepath = fmt.Sprintf("%s_%x", iface, md5.Sum([]byte(strings.Join(domains, "_"))))
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
		runEvery:        sleepDuration,
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
		"runEvery":        cfg.runEvery,
		"stateFilepath":   cfg.stateFilepath,
		"token":           fmt.Sprintf("[%d characters]", len(cfg.token)),
		"ttl":             cfg.ttl,
	}).Debug("Configuration")
}

func run(cfg runConfig) {
	addr := getIpv6Address(cfg.iface, cfg.prioritySubnets)

	if cfg.stateFilepath != "" && addr == getOldIpv6Address(cfg.stateFilepath) {
		log.Info("The address hasn't changed, nothing to do")
		log.Info("To bypass this check run without the --with-state-file flag or remove the state file: ", cfg.stateFilepath)
		return
	}

	api, err := cloudflare.NewWithAPIToken(cfg.token)
	if err != nil {
		log.WithError(err).Fatal("Couldn't create API client")
	}

	for _, domain := range cfg.domains {
		log.Info("Processing domain: ", domain)
		processDomain(api, domain, addr, cfg)
	}

	if cfg.stateFilepath != "" {
		setOldIpv6Address(cfg.stateFilepath, addr)
	}
}

func processDomain(api *cloudflare.API, domain string, addr string, cfg runConfig) {
	ctx := context.Background()

	zoneID, err := api.ZoneIDByName(getZoneFromDomain(domain))
	if err != nil {
		log.WithError(err).Fatal("Couldn't get ZoneID")
	}

	dnsRecordFilter := cloudflare.ListDNSRecordsParams{Type: "AAAA", Name: domain}
	existingDNSRecords, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), dnsRecordFilter)
	if err != nil {
		log.WithError(err).WithField("filter", dnsRecordFilter).Fatal("Couldn't get DNS records")
	}
	// If there is already a record with the same address, we want to process it
	// first. Cloudflare API doesn't allow creating multiple records with the
	// same address, which may happen in the multihost mode.
	sort.Slice(existingDNSRecords, func(i, j int) bool {
		return existingDNSRecords[i].Content == addr
	})
	for _, record := range existingDNSRecords {
		log.WithFields(log.Fields{
			"comment": record.Comment,
			"content": record.Content,
			"domain":  record.Name,
			"proxied": *record.Proxied,
			"ttl":     record.TTL,
		}).Debug("Found DNS record")
	}

	desiredDNSRecord := cloudflare.DNSRecord{Type: "AAAA", Name: domain, Content: addr, TTL: cfg.ttl}
	if cfg.multihost {
		desiredDNSRecord.Comment = cfg.hostId
	}

	// If there are no existing records, create a new one.
	if len(existingDNSRecords) == 0 {
		createNewDNSRecord(api, zoneID, desiredDNSRecord)
		return
	}

	sameHostFn := func(record cloudflare.DNSRecord, cfg runConfig) bool {
		recHost := strings.Split(record.Comment, " ")[0]
		cfgHost := strings.Split(cfg.hostId, " ")[0]
		return recHost == cfgHost
	}

	// Look through all existing records.
	// Update the matching record if found, delete the rest.
	// If no matching record is found, create a new one.
	updated := false
	for _, record := range existingDNSRecords {
		// If a record has the same address as the desired record, update it (ttl
		// or comment may have changed).
		// If a record has the same comment as the desired record and multihost is
		// enabled, update it (address or ttl may have changed).
		if !updated && (record.Content == addr || cfg.multihost && sameHostFn(record, cfg)) {
			updateDNSRecord(api, zoneID, record, desiredDNSRecord)
			updated = true
			continue
		}

		// In multihost mode, delete all records with the same host-id as the
		// current host. This should not happen during normal operation.
		if updated && record.Comment == cfg.hostId && cfg.multihost {
			log.WithField("record", record).
				Warn("Found another record with the same host-id as the current host, deleting it")
			deleteDNSRecord(api, zoneID, record)
			continue
		}

		// In single host mode, delete all other records.
		if updated && !cfg.multihost {
			deleteDNSRecord(api, zoneID, record)
			continue
		}
	}

	if !updated {
		createNewDNSRecord(api, zoneID, desiredDNSRecord)
	}
}

func createNewDNSRecord(api *cloudflare.API, zoneID string, desiredDNSRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	log.WithField("record", desiredDNSRecord).Info("Create new DNS record")
	_, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Type:    desiredDNSRecord.Type,
		Name:    desiredDNSRecord.Name,
		Content: desiredDNSRecord.Content,
		Comment: desiredDNSRecord.Comment,
		TTL:     desiredDNSRecord.TTL,
	})
	if err != nil {
		log.WithError(err).Fatal("Couldn't create DNS record")
	}
}

func updateDNSRecord(api *cloudflare.API, zoneID string, oldRecord cloudflare.DNSRecord, newRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	if oldRecord.Content == newRecord.Content &&
		oldRecord.TTL == newRecord.TTL &&
		oldRecord.Comment == newRecord.Comment {
		log.WithField("record", oldRecord).Info("DNS record is up to date")
		return
	}

	log.WithFields(log.Fields{
		"new": newRecord,
		"old": oldRecord,
	}).Info("Updating existing DNS record")
	_, err := api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
		ID:      oldRecord.ID,
		Type:    newRecord.Type,
		Name:    newRecord.Name,
		Content: newRecord.Content,
		Comment: &newRecord.Comment,
		TTL:     newRecord.TTL,
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"new": newRecord,
			"old": oldRecord,
		}).Fatal("Couldn't update DNS record")
	}
}

func deleteDNSRecord(api *cloudflare.API, zoneID string, record cloudflare.DNSRecord) {
	ctx := context.Background()
	log.WithField("record", record).Info("Deleting DNS record")
	err := api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), record.ID)
	if err != nil {
		log.WithError(err).WithField("record", record).Fatal("Couldn't delete DNS record")
	}
}

func getIpv6Address(iface string, prioritySubnets []string) string {
	netIface, err := net.InterfaceByName(iface)
	if err != nil {
		log.WithError(err).WithField("iface", iface).Fatal("Can't get the interface")
	}
	log.WithField("interface", netIface).Debug("Found the interface")

	addrs, err := netIface.Addrs()
	if err != nil {
		log.WithError(err).Fatal("Couldn't get interface addresses")
	}

	// ip.IsGlobalUnicast() returns true for:
	// GUA = Global Unicast Address
	// ULA = Unique Local Address
	// We prefer GUA over ULA.
	ipv6Addresses := []net.IP{}
	for _, addr := range addrs {
		log.WithField("address", addr).Debug("Found address")
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			log.WithError(err).WithField("address", addr).Error("Couldn't parse address")
			continue
		}
		if ip.IsGlobalUnicast() && ip.To4() == nil {
			ipv6Addresses = append(ipv6Addresses, ip)
		}
	}

	if len(ipv6Addresses) == 0 {
		log.Fatal("No suitable IPv6 addresses found")
	}

	// Sort addresses placing GUAs first
	sort.Slice(ipv6Addresses, func(i, j int) bool {
		return ipv6IsGUA(ipv6Addresses[i]) && !ipv6IsGUA(ipv6Addresses[j]) ||
			ipv6IsEUI64(ipv6Addresses[i]) && !ipv6IsEUI64(ipv6Addresses[j])
	})

	netPrioritySubnets := []net.IPNet{}
	for _, subnet := range prioritySubnets {
		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			log.WithError(err).WithField("subnet", subnet).Error("Couldn't parse subnet")
			continue
		}
		netPrioritySubnets = append(netPrioritySubnets, *ipNet)
	}

	maxWeight := len(netPrioritySubnets)
	weightedAddresses := make(map[string]int)
	for _, ip := range ipv6Addresses {
		weightedAddresses[ip.String()] = maxWeight
		for i, ipNet := range netPrioritySubnets {
			if ipNet.Contains(ip) {
				weightedAddresses[ip.String()] = i
				break
			}
		}
	}
	log.WithFields(log.Fields{
		"addresses": ipv6Addresses,
		"weighted":  weightedAddresses,
	}).Debug("Found and weighted public IPv6 addresses")

	var selectedIp string
	selectedWeight := maxWeight + 1
	for ip, weight := range weightedAddresses {
		if weight < selectedWeight {
			selectedIp = ip
			selectedWeight = weight
		}
	}

	log.WithField("addresses", ipv6Addresses).Infof("Found %d public IPv6 addresses, selected %s", len(ipv6Addresses), selectedIp)
	ipIP := net.ParseIP(selectedIp)
	if !ipv6IsEUI64(ipIP) {
		log.Warn("The selected address doesn't have a unique EUI-64, it may change frequently")
	}
	if !ipv6IsGUA(ipIP) {
		log.Warn("The selected address is not a GUA, it may not be routable")
	}
	return selectedIp
}

func getZoneFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	return strings.Join(parts[len(parts)-2:], ".")
}

func getOldIpv6Address(stateFilepath string) string {
	ipv6, err := os.ReadFile(stateFilepath)
	if err != nil {
		log.WithError(err).Warn("Can't get old ipv6 address")
		return "INVALID"
	}
	return string(ipv6)
}

func setOldIpv6Address(stateFilepath string, ipv6 string) {
	err := os.WriteFile(stateFilepath, []byte(ipv6), 0o644)
	if err != nil {
		log.WithError(err).Error("Can't write state file")
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

// Custom function to check if an IPv6 address is a GUA.
// net.IP.IsGlobalUnicast() returns true also for ULAs.
func ipv6IsGUA(ip net.IP) bool {
	return ip[0]&0b11100000 == 0b00100000
}

// Custom function to check if an IPv6 address is generated using the EUI-64 format.
// See RFC 4291, section 2.5.1.
func ipv6IsEUI64(ip net.IP) bool {
	// If the seventh bit from the left of the Interface ID is 1, and "FF FE" is
	// found in the middle of the Interface ID, then the address is generated
	// using the EUI-64 format.
	return ip[8]&0b00000010 == 0b00000010 && ip[11] == 0xff && ip[12] == 0xfe
}
