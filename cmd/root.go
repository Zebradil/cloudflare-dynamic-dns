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
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Root command invoked")
		var (
			iface  = viper.GetString("iface")
			domain = viper.GetString("domain")
			token  = viper.GetString("token")
		)
		log.WithFields(log.Fields{
			"iface":  iface,
			"domain": domain,
			"token":  fmt.Sprintf("[%d characters]", len(token)),
		}).Info("Configuration")

		addr := getIpv6Address(iface)

		api, err := cloudflare.NewWithAPIToken(token)
		if err != nil {
			log.WithError(err).Fatal("Couldn't create API client")
		}

		ctx := context.Background()

		zoneId, err := api.ZoneIDByName(getZoneFromDomain(domain))
		if err != nil {
			log.WithError(err).Fatal("Couldn't get ZoneID")
		}

		dnsRecordFilter := cloudflare.DNSRecord{Type: "AAAA", Name: domain}
		existingDnsRecords, err := api.DNSRecords(ctx, zoneId, dnsRecordFilter)
		if err != nil {
			log.WithError(err).WithField("filter", dnsRecordFilter).Fatal("Couldn't get DNS records")
		}
		log.WithField("records", existingDnsRecords).Debug("Found DNS records")

		desiredDnsRecord := cloudflare.DNSRecord{Type: "AAAA", Name: domain, Content: addr, TTL: 60}

		if len(existingDnsRecords) == 0 {
			log.WithField("record", desiredDnsRecord).Info("Create new DNS record")
			_, err := api.CreateDNSRecord(ctx, zoneId, desiredDnsRecord)
			if err != nil {
				log.WithError(err).Fatal("Couldn't create DNS record")
			}
		} else if len(existingDnsRecords) == 1 {
			// TODO do not update if there are no changes
			//      Which fields to compare?
			log.WithFields(log.Fields{
				"new": desiredDnsRecord,
				"old": existingDnsRecords[0],
			}).Info("Updating existing DNS record")
			err := api.UpdateDNSRecord(ctx, zoneId, existingDnsRecords[0].ID, desiredDnsRecord)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"new": desiredDnsRecord,
					"old": existingDnsRecords[0],
				}).Fatal("Couldn't update DNS record")
			}
		} else {
			// TODO cleanup records
			log.Fatal("Not implemented: the case when there are multiple AAAA records already")
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
	net_iface, err := net.InterfaceByName(iface)
	if err != nil {
		log.WithError(err).WithField("iface", iface).Fatal("Can't get the interface")
	}
	log.WithField("interface", net_iface).Debug("Found the interface")
	addresses, err := net_iface.Addrs()
	if err != nil {
		log.WithError(err).Fatal("Couldn't get interface addresses")
	}
	public_ipv6_addresses := []string{}
	for _, addr := range addresses {
		log.WithField("address", addr).Debug("Found address")
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() && ipnet.IP.To4() == nil {
			public_ipv6_addresses = append(public_ipv6_addresses, ipnet.IP.String())
		}
	}
	if len(public_ipv6_addresses) == 0 {
		log.Fatal("No public IPv6 addresses found")
	}
	log.WithField("addresses", public_ipv6_addresses).Infof("Found %d public IPv6 addresses, use the first one", len(public_ipv6_addresses))
	return public_ipv6_addresses[0]
}

func getZoneFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	return strings.Join(parts[len(parts)-2:], ".")
}
