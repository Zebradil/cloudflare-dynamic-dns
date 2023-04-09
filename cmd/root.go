package cmd

import (
	"context"
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cloudflare/cloudflare-go"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudflare-dynamic-dns",
	Short: "Updates AAAA records at Cloudflare according to the current IPv6 address",
	Long: `Updates AAAA records at Cloudflare according to the current IPv6 address.

Requires a network interface name for a IPv6 address lookup, domain name[s]
and Cloudflare API token with edit access rights to corresponding DNS zone.`,
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
			domains       = viper.GetStringSlice("domains")
			iface         = viper.GetString("iface")
			systemd       = viper.GetBool("systemd")
			token         = viper.GetString("token")
			ttl           = viper.GetInt("ttl")
			stateFilepath = ""
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
			"domains":       domains,
			"iface":         iface,
			"stateFilepath": stateFilepath,
			"systemd":       systemd,
			"token":         fmt.Sprintf("[%d characters]", len(token)),
			"ttl":           ttl,
		}).Info("Configuration")

		if len(domains) == 0 {
			log.Fatal("No domains specified")
		}

		addr := getIpv6Address(iface)

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
	err := api.UpdateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateDNSRecordParams{
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

	rootCmd.Flags().Bool("systemd", false, `Switch operation mode for running in systemd
In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API`)
	rootCmd.Flags().Int("ttl", 1, "Time to live, in seconds, of the DNS record. Must be between 60 and 86400, or 1 for 'automatic'")
	rootCmd.Flags().StringSlice("domains", []string{}, "Domain names to assign the IPv6 address to")
	rootCmd.Flags().String("iface", "", "Network interface to look up for a IPv6 address")
	rootCmd.Flags().String("log-level", "info", "Sets logging level: trace, debug, info, warning, error, fatal, panic")
	rootCmd.Flags().String("token", "", "Cloudflare API token with DNS edit access rights")

	viper.BindPFlags(rootCmd.Flags())
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

func getIpv6Address(iface string) string {
	netIface, err := net.InterfaceByName(iface)
	if err != nil {
		log.WithError(err).WithField("iface", iface).Fatal("Can't get the interface")
	}
	log.WithField("interface", netIface).Debug("Found the interface")
	addresses, err := netIface.Addrs()
	if err != nil {
		log.WithError(err).Fatal("Couldn't get interface addresses")
	}
	publicIpv6Addresses := []string{}
	for _, addr := range addresses {
		log.WithField("address", addr).Debug("Found address")
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() == nil {
			publicIpv6Addresses = append(publicIpv6Addresses, ipnet.IP.String())
		}
	}
	if len(publicIpv6Addresses) == 0 {
		log.Fatal("No public IPv6 addresses found")
	}
	log.WithField("addresses", publicIpv6Addresses).Infof("Found %d public IPv6 addresses, use the first one", len(publicIpv6Addresses))
	return publicIpv6Addresses[0]
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
	err := os.WriteFile(stateFilepath, []byte(ipv6), 0644)
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
	if info.Mode()&0011 != 0 {
		log.Warn("Config file should be accessible only by owner")
	}
}
