package cmd

import (
	"context"
	"sort"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/zones"
	log "github.com/sirupsen/logrus"
	"github.com/weppos/publicsuffix-go/publicsuffix"
)

func processDomain(api *cloudflare.Client, domain string, addr string, cfg runConfig) {
	ctx := context.Background()
	listType := getRecordListType(cfg.stack)

	zoneName, err := publicsuffix.Domain(domain)
	if err != nil {
		log.WithError(err).Fatal("Couldn't get ZoneName")
	} else {
		log.WithField("zoneName", zoneName).Debug("Got ZoneName")
	}

	zoneRes, err := api.Zones.List(ctx, zones.ZoneListParams{Name: cloudflare.F(zoneName)})
	if err != nil {
		log.WithError(err).Fatal("Couldn't list zones")
	}
	if len(zoneRes.Result) == 0 {
		log.WithField("zoneName", zoneName).Fatal("No zone found")
	}
	zoneID := zoneRes.Result[0].ID

	dnsRecordFilter := dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		Type:   cloudflare.F(listType),
		Name:   cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.F(domain)}),
	}
	listRes, err := api.DNS.Records.List(ctx, dnsRecordFilter)
	if err != nil {
		log.WithError(err).WithField("filter", dnsRecordFilter).Fatal("Couldn't get DNS records")
	}
	existingDNSRecords := listRes.Result

	// If there are no existing records, create a new one and exit.
	if len(existingDNSRecords) == 0 {
		createNewDNSRecord(api, zoneID, domain, addr, cfg)
		return
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
			"proxied": record.Proxied,
			"ttl":     record.TTL,
			"type":    record.Type,
		}).Debug("Found DNS record")
	}

	sameHostFn := func(record dns.RecordResponse, cfg runConfig) bool {
		recHost := strings.Split(record.Comment, " ")[0]
		cfgHost := strings.Split(cfg.hostID, " ")[0]
		return recHost == cfgHost
	}

	// Update the current record if it is either:
	//   1. has the same address as the desired record (proxied, ttl, or comment may have changed),
	//   2. has the same comment as the desired record and multihost is enabled (address and possibly proxied or ttl have changed),
	//   3. has empty comment and multihost is disabled (address and possibly proxied or ttl have changed).
	// NOTE: despite API returning empty Comment as `null`, cloudflare-go represents it as an empty string.
	//       This can break in the future.
	shouldUpdateFn := func(record dns.RecordResponse, cfg runConfig) bool {
		return record.Content == addr ||
			cfg.multihost && sameHostFn(record, cfg) ||
			!cfg.multihost && record.Comment == ""
	}

	// Look through all existing records.
	// Update the matching record if found, delete the rest.
	// If no matching record is found, create a new one.
	updated := false
	for _, record := range existingDNSRecords {
		if !updated && shouldUpdateFn(record, cfg) {
			updateDNSRecord(api, zoneID, record, addr, cfg)
			updated = true
			continue
		}

		// In multihost mode, delete all records with the same host-id as the
		// current host. This should not happen during normal operation.
		if updated && record.Comment == cfg.hostID && cfg.multihost {
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
		createNewDNSRecord(api, zoneID, domain, addr, cfg)
	}
}

func getRecordListType(stack IPStack) dns.RecordListParamsType {
	if stack == ipv4 {
		return dns.RecordListParamsTypeA
	}
	if stack == ipv6 {
		return dns.RecordListParamsTypeAAAA
	}
	log.WithField("stack", stack).Fatal("Invalid IP mode")
	return ""
}

func newRecordBody(stack IPStack, name, content string, ttl int, comment string, proxied *bool) dns.RecordNewParamsBodyUnion {
	if stack == ipv4 {
		rec := dns.ARecordParam{
			Name:    cloudflare.F(name),
			Content: cloudflare.F(content),
			Type:    cloudflare.F(dns.ARecordTypeA),
			TTL:     cloudflare.F(dns.TTL(ttl)),
			Comment: cloudflare.F(comment),
		}
		if proxied != nil {
			rec.Proxied = cloudflare.F(*proxied)
		}
		return rec
	}
	rec := dns.AAAARecordParam{
		Name:    cloudflare.F(name),
		Content: cloudflare.F(content),
		Type:    cloudflare.F(dns.AAAARecordTypeAAAA),
		TTL:     cloudflare.F(dns.TTL(ttl)),
		Comment: cloudflare.F(comment),
	}
	if proxied != nil {
		rec.Proxied = cloudflare.F(*proxied)
	}
	return rec
}

func updateRecordBody(stack IPStack, name, content string, ttl int, comment string, proxied *bool) dns.RecordUpdateParamsBodyUnion {
	if stack == ipv4 {
		rec := dns.ARecordParam{
			Name:    cloudflare.F(name),
			Content: cloudflare.F(content),
			Type:    cloudflare.F(dns.ARecordTypeA),
			TTL:     cloudflare.F(dns.TTL(ttl)),
			Comment: cloudflare.F(comment),
		}
		if proxied != nil {
			rec.Proxied = cloudflare.F(*proxied)
		}
		return rec
	}
	rec := dns.AAAARecordParam{
		Name:    cloudflare.F(name),
		Content: cloudflare.F(content),
		Type:    cloudflare.F(dns.AAAARecordTypeAAAA),
		TTL:     cloudflare.F(dns.TTL(ttl)),
		Comment: cloudflare.F(comment),
	}
	if proxied != nil {
		rec.Proxied = cloudflare.F(*proxied)
	}
	return rec
}

func createNewDNSRecord(api *cloudflare.Client, zoneID string, domain, addr string, cfg runConfig) {
	ctx := context.Background()

	var comment string
	if cfg.multihost {
		comment = cfg.hostID
	}

	var proxied *bool
	if cfg.proxy != "auto" {
		proxied = Ptr(cfg.proxy == "enabled")
	}

	log.WithFields(log.Fields{
		"domain":  domain,
		"content": addr,
	}).Info("Create new DNS record")
	_, err := api.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   newRecordBody(cfg.stack, domain, addr, cfg.ttl, comment, proxied),
	})
	if err != nil {
		log.WithError(err).Fatal("Couldn't create DNS record")
	}
}

func updateDNSRecord(api *cloudflare.Client, zoneID string, oldRecord dns.RecordResponse, newAddr string, cfg runConfig) {
	ctx := context.Background()

	var newComment string
	if cfg.multihost {
		newComment = cfg.hostID
	}

	var proxied *bool
	if cfg.proxy != "auto" {
		proxied = Ptr(cfg.proxy == "enabled")
	}

	proxiedMatches := proxied == nil || oldRecord.Proxied == *proxied
	if oldRecord.Content == newAddr &&
		proxiedMatches &&
		int(oldRecord.TTL) == cfg.ttl &&
		oldRecord.Comment == newComment {
		log.WithField("record", oldRecord).Info("DNS record is up to date")
		return
	}

	log.WithFields(log.Fields{
		"old": oldRecord,
		"new": newAddr,
	}).Info("Updating existing DNS record")

	_, err := api.DNS.Records.Update(ctx, oldRecord.ID, dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneID),
		Body:   updateRecordBody(cfg.stack, oldRecord.Name, newAddr, cfg.ttl, newComment, proxied),
	})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"old": oldRecord,
			"new": newAddr,
		}).Fatal("Couldn't update DNS record")
	}
}

func deleteDNSRecord(api *cloudflare.Client, zoneID string, record dns.RecordResponse) {
	ctx := context.Background()
	log.WithField("record", record).Info("Deleting DNS record")
	_, err := api.DNS.Records.Delete(ctx, record.ID, dns.RecordDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		log.WithError(err).WithField("record", record).Fatal("Couldn't delete DNS record")
	}
}
