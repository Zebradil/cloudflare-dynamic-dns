package cmd

import (
	"net"
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
)

type ipStack interface {
	filterIPs([]net.IP) []net.IP
	logIP(net.IP)
	sortIPs([]net.IP) []net.IP
}

type ipManager struct {
	cfg runConfig
	ipStack
}

type ipv4Stack struct{}

type ipv6Stack struct{}

func newIpManager(cfg runConfig) ipManager {
	switch cfg.stack {
	case ipv4:
		return ipManager{cfg, ipv4Stack{}}
	case ipv6:
		return ipManager{cfg, ipv6Stack{}}
	default:
		log.WithField("stack", cfg.stack).Fatal("Unknown stack")
		return ipManager{}
	}
}

func (mgr ipManager) getIP() string {
	ips := mgr.getAllIPs()
	ips = mgr.ipStack.filterIPs(ips)
	if len(ips) == 0 {
		log.Fatal("No suitable addresses found")
	}
	ips = mgr.ipStack.sortIPs(ips)
	ip := mgr.pickIP(ips)
	log.WithField("addresses", ips).Infof("Found %d public IPv6 addresses, selected %s", len(ips), ip)
	mgr.ipStack.logIP(ip)
	return ip.String()
}

func (s ipManager) getAllIPs() []net.IP {
	iface := s.cfg.iface
	netIface, err := net.InterfaceByName(iface)
	if err != nil {
		log.WithError(err).WithField("iface", iface).Fatal("Can't get the interface")
	}
	log.WithField("interface", netIface).Debug("Found the interface")

	addrs, err := netIface.Addrs()
	if err != nil {
		log.WithError(err).Fatal("Couldn't get interface addresses")
	}

	ips := []net.IP{}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			log.WithError(err).WithField("address", addr).Error("Couldn't parse address")
			continue
		}
		ips = append(ips, ip)
	}
	return ips
}

func (s ipv6Stack) filterIPs(ips []net.IP) []net.IP {
	// ip.IsGlobalUnicast() returns true for:
	// GUA = Global Unicast Address
	// ULA = Unique Local Address
	// We prefer GUA over ULA.
	ipv6s := []net.IP{}
	for _, ip := range ips {
		if ip.IsGlobalUnicast() && ip.To4() == nil {
			ipv6s = append(ipv6s, ip)
		}
	}

	return ipv6s
}

func (s ipv4Stack) filterIPs(ips []net.IP) []net.IP {
	ipv4s := []net.IP{}
	for _, ip := range ips {
		if ip.To4() != nil {
			ipv4s = append(ipv4s, ip)
		}
	}

	return ipv4s
}

func (s ipv6Stack) sortIPs(ips []net.IP) []net.IP {
	// Sort addresses placing GUAs first
	sort.Slice(ips, func(i, j int) bool {
		return ipv6IsGUA(ips[i]) && !ipv6IsGUA(ips[j]) ||
			ipv6IsEUI64(ips[i]) && !ipv6IsEUI64(ips[j])
	})
	return ips
}

func (s ipv4Stack) sortIPs(ips []net.IP) []net.IP {
	// Sort addresses placing GUAs first, then Shared Address Space addresses,
	// then private addresses, then loopback addresses.
	sort.Slice(ips, func(i, j int) bool {
		return ips[i].IsGlobalUnicast() && !ips[j].IsGlobalUnicast() ||
			ipv4IsSAS(ips[i]) && !ipv4IsSAS(ips[j]) ||
			ips[i].IsPrivate() && !ips[j].IsPrivate()
	})
	return ips
}

func (s ipManager) pickIP(ips []net.IP) net.IP {
	netPrioritySubnets := []net.IPNet{}
	for _, subnet := range s.cfg.prioritySubnets {
		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			log.WithError(err).WithField("subnet", subnet).Error("Couldn't parse subnet")
			continue
		}
		netPrioritySubnets = append(netPrioritySubnets, *ipNet)
	}

	maxWeight := len(netPrioritySubnets)
	weightedAddresses := make(map[string]int)
	for _, ip := range ips {
		weightedAddresses[ip.String()] = maxWeight
		for i, ipNet := range netPrioritySubnets {
			if ipNet.Contains(ip) {
				weightedAddresses[ip.String()] = i
				break
			}
		}
	}
	log.WithFields(log.Fields{
		"addresses": ips,
		"weighted":  weightedAddresses,
	}).Debug("Found and weighted public IP addresses")

	var selectedIp string
	selectedWeight := maxWeight + 1
	for ip, weight := range weightedAddresses {
		if weight < selectedWeight {
			selectedIp = ip
			selectedWeight = weight
		}
	}

	return net.ParseIP(selectedIp)
}

func (s ipv6Stack) logIP(ip net.IP) {
	if !ipv6IsEUI64(ip) {
		log.Warn("The selected address doesn't have a unique EUI-64, it may change frequently")
	}
	if !ipv6IsGUA(ip) {
		log.Warn("The selected address is not a GUA, it may not be routable")
	}
}

func (s ipv4Stack) logIP(ip net.IP) {
	if ipv4IsSAS(ip) {
		log.Warn("The selected address is in the Shared Address Space range")
	}
	if ip.IsPrivate() {
		log.Warn("The selected address is private, it may not be routable")
	}
	if ip.IsLoopback() {
		log.Warn("The selected address is a loopback address")
	}
}

func (mgr ipManager) getOldIp() string {
	ip, err := os.ReadFile(mgr.cfg.stateFilepath)
	if err != nil {
		log.WithError(err).Warn("Can't get old ipv6 address")
		return "INVALID"
	}
	return string(ip)
}

func (mgr ipManager) setOldIp(ip string) {
	err := os.WriteFile(mgr.cfg.stateFilepath, []byte(ip), 0o644)
	if err != nil {
		log.WithError(err).Error("Can't write state file")
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

func ipv4IsSAS(ip net.IP) bool {
	// The Shared Address Space address range is 100.64.0.0/10.
	// See RFC 6598.
	return ip[0] == 100 && ip[1] >= 64 && ip[1] < 128
}
