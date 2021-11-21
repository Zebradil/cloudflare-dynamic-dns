/*
Copyright © 2021 German Lashevich <german.lashevich@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
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

Requires a network interface name for a IPv6 address lookup, domain name
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
			iface         = viper.GetString("iface")
			domain        = viper.GetString("domain")
			token         = viper.GetString("token")
			systemd       = viper.GetBool("systemd")
			stateFilepath = ""
		)

		if systemd {
			stateFilepath = filepath.Join(os.Getenv("STATE_DIRECTORY"), domain)
		}

		log.WithFields(log.Fields{
			"iface":         iface,
			"domain":        domain,
			"token":         fmt.Sprintf("[%d characters]", len(token)),
			"systemd":       systemd,
			"stateFilepath": stateFilepath,
		}).Info("Configuration")

		addr := getIpv6Address(iface)

		if systemd && addr == getOldIpv6Address(stateFilepath) {
			log.Info("The address hasn't changed, nothing to do")
			log.Info(fmt.Sprintf("To bypass this check run without --systemd flag or remove the state file: %s", stateFilepath))
			return
		}

		api, err := cloudflare.NewWithAPIToken(token)
		if err != nil {
			log.WithError(err).Fatal("Couldn't create API client")
		}

		ctx := context.Background()

		zoneID, err := api.ZoneIDByName(getZoneFromDomain(domain))
		if err != nil {
			log.WithError(err).Fatal("Couldn't get ZoneID")
		}

		dnsRecordFilter := cloudflare.DNSRecord{Type: "AAAA", Name: domain}
		existingDNSRecords, err := api.DNSRecords(ctx, zoneID, dnsRecordFilter)
		if err != nil {
			log.WithError(err).WithField("filter", dnsRecordFilter).Fatal("Couldn't get DNS records")
		}
		log.WithField("records", existingDNSRecords).Debug("Found DNS records")

		desiredDNSRecord := cloudflare.DNSRecord{Type: "AAAA", Name: domain, Content: addr, TTL: 60}

		if len(existingDNSRecords) == 0 {
			log.WithField("record", desiredDNSRecord).Info("Create new DNS record")
			_, err := api.CreateDNSRecord(ctx, zoneID, desiredDNSRecord)
			if err != nil {
				log.WithError(err).Fatal("Couldn't create DNS record")
			}
		} else if len(existingDNSRecords) == 1 {
			if existingDNSRecords[0].Content == desiredDNSRecord.Content && existingDNSRecords[0].TTL ==
				desiredDNSRecord.TTL {
				log.WithField("record", existingDNSRecords[0]).Info("DNS record is up to date")
				return
			}

			log.WithFields(log.Fields{
				"new": desiredDNSRecord,
				"old": existingDNSRecords[0],
			}).Info("Updating existing DNS record")
			err := api.UpdateDNSRecord(ctx, zoneID, existingDNSRecords[0].ID, desiredDNSRecord)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"new": desiredDNSRecord,
					"old": existingDNSRecords[0],
				}).Fatal("Couldn't update DNS record")
			}
		} else {
			// TODO cleanup records
			log.Fatal("Not implemented: the case when there are multiple AAAA records already")
		}
		if systemd {
			setOldIpv6Address(stateFilepath, addr)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cloudflare-dynamic-dns.yaml)")

	rootCmd.Flags().String("iface", "", "Network interface to look up for a IPv6 address")
	rootCmd.Flags().String("domain", "", "Domain name to assign the IPv6 address to")
	rootCmd.Flags().String("token", "", "Cloudflare API token with DNS edit access rights")
	rootCmd.Flags().String("log-level", "info", "Sets logging level: trace, debug, info, warning, error, fatal, panic")
	rootCmd.Flags().Bool("systemd", false, `Switch operation mode for running in systemd
In this mode previously used ipv6 address is preserved between runs to avoid unnecessary calls to CloudFlare API`)

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
	if info.Mode()&1010 != 0 {
		log.Warn("Config file should be accessible only by owner")
	}
}
