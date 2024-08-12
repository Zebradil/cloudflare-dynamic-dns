package cmd

import (
	"net"
	"os"

	log "github.com/sirupsen/logrus"
)

type ipStack interface {
	logIP(net.IP)
	getBaseScore(net.IP) uint16
}

type ipManager struct {
	cfg runConfig
	ipStack
}

type ipv4Stack struct{}

type ipv6Stack struct{}

func newIPManager(cfg runConfig) ipManager {
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
	ip := mgr.pickIP(ips)
	if ip == nil {
		log.Fatal("No suitable addresses found")
	}
	log.WithField("addresses", ips).Infof("Found %d public IP addresses, selected %s", len(ips), ip)
	mgr.logIP(ip)
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

func (s ipv6Stack) getBaseScore(ip net.IP) uint16 {
	score := uint16(1)

	if ip.To4() != nil {
		return 0
	}
	if !ip.IsGlobalUnicast() {
		return 0
	}

	if ipv6IsGUA(ip) {
		score += 0x8
	}

	if ipv6IsEUI64(ip) {
		score += 0x4
	}

	return score
}

func (s ipv4Stack) getBaseScore(ip net.IP) uint16 {
	score := uint16(1)

	if ip.To4() == nil {
		return 0
	}
	if ip.IsGlobalUnicast() {
		score += 0x8
	}
	if ipv4IsSAS(ip) {
		score += 0x4
	}
	if ip.IsPrivate() {
		score += 0x2
	}

	return score
}

// The score function evaluates the value of a given IP address and returns
// a score of type uint64.
// The higher the score, the more valuable the IP address.
func (s ipManager) score(ip net.IP) uint64 {
	// Score format:
	// +------------+------------------+--------------+
	// | Reserved   | Priority Subnets | Address Type |
	// | (16 bits)  |    (32 bits)     |   (16 bits)  |
	// +------------+------------------+--------------+
	score := uint64(s.getBaseScore(ip))

	// Scoring by priority subnet
	for order, subnet := range s.cfg.prioritySubnets {
		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			log.WithError(err).WithField("subnet", subnet).Error("Couldn't parse subnet")
			continue
		}
		if ipNet.Contains(ip) {
			score += uint64(len(s.cfg.prioritySubnets) - order) << 16
			break
		}
	}
	return score
}

func (s ipManager) pickIP(ips []net.IP) net.IP {
	bestIpIdx := -1
	bestScore := uint64(0) // any address with score 0 will not be picked
	for idx, ip := range ips {
		score := s.score(ip)
		log.WithFields(log.Fields{
			"address": ip,
			"score":   score,
		}).Debug("Address score")
		if score > bestScore {
			bestIpIdx = idx
			bestScore = score
		}
	}
	if bestIpIdx < 0 {
		return nil
	}
	return ips[bestIpIdx]
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

func (mgr ipManager) getOldIP() string {
	ip, err := os.ReadFile(mgr.cfg.stateFilepath)
	if err != nil {
		log.WithError(err).Warn("Can't get old ipv6 address")
		return "INVALID"
	}
	return string(ip)
}

func (mgr ipManager) setOldIP(ip string) {
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
