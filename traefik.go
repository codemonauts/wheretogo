package main

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

func updateTraefikConfig(c *Config) error {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return err
	}

	kv := client.KV()
	ops := api.KVTxnOps{}

	for _, entry := range c.Entrys {
		ops = append(ops, &api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   fmt.Sprintf("traefik/frontends/%s/backend", entry.Domain),
			Value: []byte("wheretogo")},
			&api.KVTxnOp{
				Verb:  api.KVSet,
				Key:   fmt.Sprintf("traefik/frontends/%s/routes/default/rule", entry.Domain),
				Value: []byte(fmt.Sprintf("Host: %s", entry.Domain)),
			})
	}

	log.Info("Writing %d keys to Consul\n", len(ops))
	log.Debug(ops)

	_, _, _, err = kv.Txn(ops, nil)
	if err != nil {
		return err
	}

	return nil
}
