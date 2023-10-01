package main

import (
	"context"

	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/client/utils"
	rendezvous "github.com/gfleury/solo/rendezvous/node"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
)

func main() {
	var logger = logger.New(log.LevelDebug)

	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("rendezvous", "debug")
	log.SetLogLevel("main", "debug")

	logger.Info("create host")

	r, err := rendezvous.NewRendezvousHost(
		context.Background(),
		logger,
		"rendezvous",
		node.ListenAddrs(false, rendezvous.DEFAULT_RENDEZVOUS_BASE_PORT),
		libp2p.AddrsFactory(utils.DefaultAddrsFactory))
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

	select {}
}
