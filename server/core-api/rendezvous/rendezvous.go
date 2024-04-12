package rendezvous

import (
	"context"

	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/node"
	"github.com/gfleury/solo/client/utils"
	"github.com/gfleury/solo/common"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	gostream "github.com/libp2p/go-libp2p-gostream"
)

func StartRendezvous(ctx context.Context, randomPort bool) (*RendezvousHost, error) {
	var logger = logger.New(log.LevelDebug)

	log.SetAllLoggers(log.LevelInfo)
	log.SetLogLevel("rendezvous", "debug")
	log.SetLogLevel("main", "debug")

	logger.Info("create host")

	r, err := NewRendezvousHost(
		ctx,
		logger,
		"rendezvous",
		node.ListenAddrs(randomPort, DEFAULT_RENDEZVOUS_BASE_PORT),
		libp2p.AddrsFactory(utils.DefaultAddrsFactory))
	if err != nil {
		return nil, err
	}

	addrs, err := r.GetAddrs()
	if err != nil {
		return nil, err
	}

	logger.Infof("Rendezvous endpoints: %s", addrs)

	err = r.Start()
	if err != nil {
		return nil, err
	}

	r.HTTPListener, err = gostream.Listen(r.host, common.SoloAPIP2PProtocol)

	return r, err
}
