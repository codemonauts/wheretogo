package main

import (
	"fmt"

	"github.com/gorilla/mux"
)

func buildRouter(c *Config) *mux.Router {
	r := mux.NewRouter()
	r.Use(loggingMiddleware)

	for _, e := range c.Entrys {
		fmt.Printf("> Loading rule: %q\n", e.Name)
		s := r.Host(e.Domain).Subrouter()
		switch e.Type {
		case "regex":
			s.PathPrefix(e.Path).HandlerFunc(getRegExHandler(e.Data)).Name(e.Name)
		case "webhook":
			s.PathPrefix(e.Path).HandlerFunc(getWebhookHandler(e.Data)).Name(e.Name)
		}
	}

	return r
}
