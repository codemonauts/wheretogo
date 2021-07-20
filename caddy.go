package main

import (
	"github.com/imroc/req"
	log "github.com/sirupsen/logrus"
)

const (
	CADDY_API = "http://localhost:2019/config/apps/http/servers/srv0/routes/0/match/0/host/..."
)

type DomainList []string

func (list DomainList) contains(s string) bool {
	for _, entry := range list {
		if entry == s {
			return true
		}
	}
	return false
}

func configureCaddy(c *Config) {
	log.Info("Configuring Caddy webserver")

	// Domains that are already configured in Caddy
	existingDomains := DomainList{}
	// The new domains we found in the config file which need to be added
	newDomains := DomainList{}

	r, err := req.Get(CADDY_API)
	if err != nil {
		log.Fatalf("Couldn't talk to Caddy API: %s", err)
	}
	r.ToJSON(&existingDomains)

	for domain := range *c {
		if !existingDomains.contains(domain) && !newDomains.contains(domain) {
			newDomains = append(newDomains, domain)
		}
	}

	r, _ = req.Post(CADDY_API, req.BodyJSON(&newDomains))
	if r.Response().StatusCode != 200 {
		log.Panic("Could not configure all new domains in Caddy")
	}
}
