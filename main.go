package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/zebradil/cloudflare-dynamic-dns/cmd"
)

func main() {
	log.SetLevel(log.DebugLevel)
	cmd.Execute()
}
