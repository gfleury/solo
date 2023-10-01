package rendezvous

import (
	"context"
	"testing"

	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/client/utils"
	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
)

func TestRendzvousNode(t *testing.T) {
	var logger = log.Logger("main")

	log.SetAllLoggers(log.LevelWarn)
	log.SetLogLevel("rendezvous", "debug")
	log.SetLogLevel("main", "debug")
	logger.Info("create host")

	r, err := NewRendezvousHost(
		context.Background(),
		"rendezvous",
		node.ListenAddrs(false, DEFAULT_RENDEZVOUS_BASE_PORT),
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
