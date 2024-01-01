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

	cloudflare "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const longDescription = `
Selects an IPv6 address from the specified network interface and updates
AAAA records at Cloudflare for configured domains.

Requires a network interface name for a IPv6 address lookup, domain name[s]
and Cloudflare API token with edit access rights to corresponding DNS zone.

When multiple IPv6 addresses are found on the interface, the following rules
are used to select the one to use:
    1. Only global unicast addresses (GUA) and unique local addresses (ULA) are
       considered.
    2. GUA addresses are preferred over ULA addresses.
    3. Unique EUI-64 addresses are preferred over randomly generated addresses.
    4. If priority subnets are specified, addresses from the subnet with the
       highest priority are selected. The priority is determined by the order
       of subnets specified on the command line or in the config file.

The program can be run in systemd mode, in which case the previously used
IPv6 address is preserved between runs to avoid unnecessary calls to Cloudflare
API. This mode is enabled by passing --systemd flag. The state file is stored
in the directory specified by the STATE_DIRECTORY environment variable.

The program can be configured using a config file. The default location is
$HOME/.cloudflare-dynamic-dns.yaml. The config file location can be overridden
using --config flag. The config file format is YAML. The following options are
supported (with example values):

    iface: eth0
    token: cloudflare-api-token
    domains:
      - example.com
      - "*.example.com"
    # --- optional fields ---
    prioritySubnets:
      - 2001:db8::/32
      - 2001:db8:1::/48
    ttl: 180
    systemd: false
`

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudflare-dynamic-dns",
	Short: "Updates AAAA records at Cloudflare according to the current IPv6 address",
	Long:  longDescription,
	Args:  cobra.NoArgs,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := log.ParseLevel(viper.GetString("log-level"))
		if err != nil {
			return err
		}
		log.Info("Setting log level to:", level)
		log.SetLevel(level)
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if viper.ConfigFileUsed() != "" {
			log.WithField("config", viper.ConfigFileUsed()).Debug("Using config file")
			checkConfigAccessMode(viper.ConfigFileUsed())
		} else {
			log.Debug("No config file used")
		}

		var (
			domains         = viper.GetStringSlice("domains")
			iface           = viper.GetString("iface")
			prioritySubnets = viper.GetStringSlice("prioritySubnets")
			stateFilepath   = ""
			systemd         = viper.GetBool("systemd")
			token           = viper.GetString("token")
			ttl             = viper.GetInt("ttl")
		)

		if ttl < 60 || ttl > 86400 {
			// NOTE: 1 is a special value which means "use the default TTL"
			if ttl != 1 {
				log.WithFields(log.Fields{"ttl": ttl}).Warn("TTL must be between 60 and 86400; using Cloudflare's default")
				ttl = 1
			}
		}

		if systemd {
			stateFilepath = filepath.Join(os.Getenv("STATE_DIRECTORY"), fmt.Sprintf("%s_%x", iface, md5.Sum([]byte(strings.Join(domains, "_")))))
		}

		log.WithFields(log.Fields{
			"domains":         domains,
			"iface":           iface,
			"prioritySubnets": prioritySubnets,
			"stateFilepath":   stateFilepath,
			"systemd":         systemd,
			"token":           fmt.Sprintf("[%d characters]", len(token)),
			"ttl":             ttl,
		}).Info("Configuration")

		if len(domains) == 0 {
			log.Fatal("No domains specified")
		}

		addr := getIpv6Address(iface, prioritySubnets)

		if systemd && addr == getOldIpv6Address(stateFilepath) {
			log.Info("The address hasn't changed, nothing to do")
			log.Info("To bypass this check run without --systemd flag or remove the state file: ", stateFilepath)
			return
		}

		api, err := cloudflare.NewWithAPIToken(token)
		if err != nil {
			log.WithError(err).Fatal("Couldn't create API client")
		}

		for _, domain := range domains {
			log.Info("Processing domain: ", domain)
			processDomain(api, domain, addr, ttl)
		}

		if systemd {
			setOldIpv6Address(stateFilepath, addr)
		}
	},
}

func processDomain(api *cloudflare.API, domain string, addr string, ttl int) {
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
	log.WithField("records", existingDNSRecords).Debug("Found DNS records")

	desiredDNSRecord := cloudflare.DNSRecord{Type: "AAAA", Name: domain, Content: addr, TTL: ttl}

	if len(existingDNSRecords) == 0 {
		createNewDNSRecord(api, zoneID, desiredDNSRecord)
	} else if len(existingDNSRecords) == 1 {
		updateDNSRecord(api, zoneID, existingDNSRecords[0], desiredDNSRecord)
	} else {
		updated := false
		for oldRecord := range existingDNSRecords {
			if !updated && existingDNSRecords[oldRecord].Content == desiredDNSRecord.Content {
				updateDNSRecord(api, zoneID, existingDNSRecords[oldRecord], desiredDNSRecord)
				updated = true
			} else {
				deleteDNSRecord(api, zoneID, existingDNSRecords[oldRecord])
			}
		}
		if !updated {
			createNewDNSRecord(api, zoneID, desiredDNSRecord)
		}
	}
}

func createNewDNSRecord(api *cloudflare.API, zoneID string, desiredDNSRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	log.WithField("record", desiredDNSRecord).Info("Create new DNS record")
	_, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Type:    desiredDNSRecord.Type,
		Name:    desiredDNSRecord.Name,
		Content: desiredDNSRecord.Content,
		TTL:     desiredDNSRecord.TTL,
	})
	if err != nil {
		log.WithError(err).Fatal("Couldn't create DNS record")
	}
}

func updateDNSRecord(api *cloudflare.API, zoneID string, oldRecord cloudflare.DNSRecord, newRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	if oldRecord.Content == newRecord.Content && oldRecord.TTL ==
		newRecord.TTL {
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-dynamic-dns.yaml)")

	rootCmd.Flags().Bool("systemd", false, `Switch operation mode for running in systemd.
In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API.`)
	rootCmd.Flags().Int("ttl", 1, "Time to live, in seconds, of the DNS record. Must be between 60 and 86400, or 1 for 'automatic'.")
	rootCmd.Flags().StringSlice("domains", []string{}, "Domain names to assign the IPv6 address to.")
	rootCmd.Flags().StringSlice("priority-subnets", []string{}, `IPv6 subnets to prefer over others.
If multiple IPv6 addresses are found on the interface, the one from the subnet with the highest priority is used.`)
	rootCmd.Flags().String("iface", "", "Network interface to look up for a IPv6 address.")
	rootCmd.Flags().String("log-level", "info", "Sets logging level: trace, debug, info, warning, error, fatal, panic.")
	rootCmd.Flags().String("token", "", "Cloudflare API token with DNS edit access rights.")

	err := viper.BindPFlags(rootCmd.Flags())
	if err != nil {
		log.WithError(err).Fatal("Couldn't bind flags")
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cloudflare-dynamic-dns" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cloudflare-dynamic-dns")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file:", viper.ConfigFileUsed())
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
			ipv6IsLocalUnique(ipv6Addresses[i]) && !ipv6IsLocalUnique(ipv6Addresses[j])
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

// Custom function to check if an EUI-64 of an IPv6 address is locally unique.
// See RFC 4291, section 2.5.1.
func ipv6IsLocalUnique(ip net.IP) bool {
	return ip[8]&0b01000000 == 0b01000000
}
