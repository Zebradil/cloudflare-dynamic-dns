package cmd

import (
	"context"
	"sort"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	log "github.com/sirupsen/logrus"
)

func processDomain(api *cloudflare.API, domain string, addr string, cfg runConfig) {
	ctx := context.Background()
	recordType := getRecordType(cfg.stack)

	zoneID, err := api.ZoneIDByName(getZoneFromDomain(domain))
	if err != nil {
		log.WithError(err).Fatal("Couldn't get ZoneID")
	}

	desiredDNSRecord := cloudflare.DNSRecord{Type: recordType, Name: domain, Content: addr, TTL: cfg.ttl}
	if cfg.multihost {
		desiredDNSRecord.Comment = cfg.hostId
	}

	if cfg.proxy != "auto" {
		desiredDNSRecord.Proxied = Ptr(cfg.proxy == "enabled")
	}

	dnsRecordFilter := cloudflare.ListDNSRecordsParams{Type: recordType, Name: domain}
	existingDNSRecords, _, err := api.ListDNSRecords(ctx, cloudflare.ZoneIdentifier(zoneID), dnsRecordFilter)
	if err != nil {
		log.WithError(err).WithField("filter", dnsRecordFilter).Fatal("Couldn't get DNS records")
	}

	// If there are no existing records, create a new one and exit.
	if len(existingDNSRecords) == 0 {
		createNewDNSRecord(api, zoneID, desiredDNSRecord)
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
			"proxied": *record.Proxied,
			"ttl":     record.TTL,
			"type":    record.Type,
		}).Debug("Found DNS record")
	}

	sameHostFn := func(record cloudflare.DNSRecord, cfg runConfig) bool {
		recHost := strings.Split(record.Comment, " ")[0]
		cfgHost := strings.Split(cfg.hostId, " ")[0]
		return recHost == cfgHost
	}

	// Update the current record if it is either:
	//   1. has the same address as the desired record (proxied, ttl, or comment may have changed),
	//   2. has the same comment as the desired record and multihost is enabled (address and possibly proxied or ttl have changed),
	//   3. has empty comment and multihost is disabled (address and possibly proxied or ttl have changed).
	// NOTE: despite API returning empty Comment as `null`, cloudflare-go represents it as an empty string.
	//       This can break in the future.
	shouldUpdateFn := func(record cloudflare.DNSRecord, cfg runConfig) bool {
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

func getRecordType(stack stack) string {
	if stack == ipv4 {
		return "A"
	}
	if stack == ipv6 {
		return "AAAA"
	}
	log.WithField("stack", stack).Fatal("Invalid IP mode")
	return ""
}

func createNewDNSRecord(api *cloudflare.API, zoneID string, desiredDNSRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	log.WithField("record", desiredDNSRecord).Info("Create new DNS record")
	_, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
		Comment: desiredDNSRecord.Comment,
		Content: desiredDNSRecord.Content,
		Name:    desiredDNSRecord.Name,
		Proxied: desiredDNSRecord.Proxied,
		TTL:     desiredDNSRecord.TTL,
		Type:    desiredDNSRecord.Type,
	})
	if err != nil {
		log.WithError(err).Fatal("Couldn't create DNS record")
	}
}

func updateDNSRecord(api *cloudflare.API, zoneID string, oldRecord cloudflare.DNSRecord, newRecord cloudflare.DNSRecord) {
	ctx := context.Background()
	proxiedMatches := newRecord.Proxied == nil ||
		(oldRecord.Proxied != nil && *oldRecord.Proxied == *newRecord.Proxied)
	if oldRecord.Content == newRecord.Content &&
		proxiedMatches &&
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
		Comment: &newRecord.Comment,
		Content: newRecord.Content,
		ID:      oldRecord.ID,
		Name:    newRecord.Name,
		Proxied: newRecord.Proxied,
		TTL:     newRecord.TTL,
		Type:    newRecord.Type,
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

func getZoneFromDomain(domain string) string {
	parts := strings.Split(domain, ".")
	return strings.Join(parts[len(parts)-2:], ".")
}
