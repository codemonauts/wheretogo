package main

import (
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func valueOrDefault(value string, defaultValue string) string {
	if value != "" {
		return value
	} else {
		return defaultValue
	}
}

func buildRouter(c *Config) *mux.Router {
	r := mux.NewRouter()

	for domain, ruleList := range *c {
		for _, c := range ruleList {
			prefix := valueOrDefault(c.PathPrefix, "/")
			match := valueOrDefault(c.Match, "(.*)")

			log.WithFields(log.Fields{
				"Host":       domain,
				"PathPrefix": prefix,
				"Match":      match,
				"Target":     c.Target,
			}).Debug("Loading regexp-rule")

			r.Host(domain).PathPrefix(prefix).HandlerFunc(getRegExHandler(match, c.Target)).Name(c.Name)
		}
	}

	r.Use(loggingMiddleware)

	return r
}
