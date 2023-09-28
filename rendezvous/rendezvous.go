package main

import (
	"context"

	rendezvous "github.com/gfleury/solo/rendezvous/node"
	"github.com/ipfs/go-log/v2"
)

func main() {
	var logger = log.Logger("main")

	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("rendezvous", "debug")
	log.SetLogLevel("main", "debug")
	logger.Info("create host")

	r, err := rendezvous.NewRendezvousHost(context.Background(), "rendezvous")
	if err != nil {
		panic(err)
	}

	addrs, err := r.GetAddrs()
	if err != nil {
		panic(err)
	}

	logger.Infof("Rendezvous endpoints: %s", addrs)

	err = r.Start()
	if err != nil {
		panic(err)
	}

	err = r.Start()
	if err != nil {
		panic(err)
	}

	select {}
}
