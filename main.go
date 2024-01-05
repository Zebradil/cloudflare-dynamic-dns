package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/zebradil/cloudflare-dynamic-dns/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetLevel(log.DebugLevel)
	cmd := cmd.NewRootCmd(version, commit, date)
	if err := cmd.Execute(); err != nil {
		log.WithError(err).Fatal("Failed to execute command")
	}
}
