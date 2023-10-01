/*
Copyright © 2021-2022 Ettore Di Giacinto <mudler@mocaccino.org>
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package node

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/conngater"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/identify"
	"github.com/libp2p/go-libp2p/p2p/security/noise"

	"github.com/gfleury/solo/client/broadcast"
	"github.com/gfleury/solo/client/config"
	"github.com/gfleury/solo/client/crypto"
	discovery "github.com/gfleury/solo/client/discovery"
	"github.com/gfleury/solo/client/logger"
	"github.com/gfleury/solo/client/utils"
	"github.com/gfleury/solo/client/vpn"
)

type Node struct {
	config      Config
	Broadcaster broadcast.Broadcaster

	host host.Host
	cg   *conngater.BasicConnectionGater
	sync.Mutex
}

var defaultLibp2pOptions = []libp2p.Option{
	libp2p.EnableNATService(),
	libp2p.EnableRelayService(),
	libp2p.EnableRelay(),
}

func NewWithConfig(cliConfig config.Config) (*Node, error) {
	lvl, err := log.LevelFromString(cliConfig.LogLevel)
	if err != nil {
		lvl = log.LevelError
	}
	logger := logger.New(lvl)

	if cliConfig.Libp2pLogLevel != "" {
		if strings.Contains(cliConfig.Libp2pLogLevel, ":") {
			logCfg := strings.Split(cliConfig.Libp2pLogLevel, ":")
			err = log.SetLogLevel(logCfg[0], logCfg[1])
			if err != nil {
				return nil, err
			}
		} else {
			lvl, err := log.LevelFromString(cliConfig.LogLevel)
			if err != nil {
				return nil, err
			}
			log.SetAllLoggers(lvl)
		}
	}

	connectionCfg, err := YAMLConnectionConfigFromToken(cliConfig.Token)
	if err != nil {
		return nil, err
	}

	if cliConfig.PublicDiscoveryPeers {
		cliConfig.DiscoveryPeers = []string{}
		for _, peer := range dht.DefaultBootstrapPeers {
			cliConfig.DiscoveryPeers = append(cliConfig.DiscoveryPeers, peer.String())
		}
	}

	discoveryPeers := config.Peers2List(cliConfig.DiscoveryPeers)

	// Configure DHT Discovery
	dhtOpts := []dht.Option{}
	dhtService := discovery.NewDHT(dhtOpts...)
	dhtService.DiscoveryInterval = time.Duration(cliConfig.DiscoveryInterval) * time.Second
	dhtService.OTPKey = connectionCfg.DiscoveryKey
	dhtService.DiscoveryPeers = discoveryPeers

	// Configure VPN
	vpnService := vpn.VPNNetworkService(vpn.InterfaceConfig{
		InterfaceMTU:     cliConfig.InterfaceMTU,
		InterfaceName:    cliConfig.InterfaceName,
		InterfaceAddress: cliConfig.InterfaceAddress,
		CreateInterface:  cliConfig.CreateInterface,
		PreSharedKey:     connectionCfg.VPNPreSharedKey,
	})

	// Configure LibP2P

	cm, err := connmgr.NewConnManager(
		1,
		cliConfig.MaxConnections,
		connmgr.WithGracePeriod(80*time.Second),
	)
	if err != nil {
		logger.Fatal("could not create connection manager")
	}

	libp2pOpts := []libp2p.Option{
		libp2p.UserAgent("solo"),
		libp2p.Security(noise.ID, noise.New),
		libp2p.ConnectionManager(cm),
	}

	libp2pOpts = append(libp2pOpts, defaultLibp2pOptions...)

	var limiter rcmgr.Limiter

	defaults := rcmgr.DefaultLimits
	def := &defaults

	libp2p.SetDefaultServiceLimits(def)
	limiter = rcmgr.NewFixedLimiter(def.AutoScale())

	rc, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		logger.Fatal("could not create resource manager")
	}

	libp2pOpts = append(libp2pOpts, libp2p.ResourceManager(rc))

	if cliConfig.HolePunch {
		libp2pOpts = append(libp2pOpts, libp2p.EnableHolePunching())
	}

	// Enable auto-relay, for behind NAT clients
	libp2pOpts = append(libp2pOpts, libp2p.EnableAutoRelayWithPeerSource(dhtService.FindClosePeers(logger)))
	pi, err := peer.AddrInfoFromP2pAddr(discoveryPeers[0])
	if err != nil {
		return nil, err
	}
	libp2pOpts = append(libp2pOpts, libp2p.EnableAutoRelayWithStaticRelays([]peer.AddrInfo{*pi}))

	// Use default addrsFactory to filter listenAddresses
	addrsFactory := libp2p.AddrsFactory(utils.DefaultAddrsFactory)
	libp2pOpts = append(libp2pOpts, addrsFactory)

	// Force holepunch to activate easily and faster after seen only once
	identify.ActivationThresh = 1
	if runtime.GOOS == "darwin" {
		identify.ActivationThresh = 0
		libp2pOpts = append(libp2pOpts, libp2p.ForceReachabilityPrivate())
	}

	nodeConfig := Config{
		BroadcastKey:      connectionCfg.BroadcastKey,
		ListenAddresses:   []discovery.AddrList{},
		RandomIdentity:    cliConfig.RandomIdentity,
		RandomPort:        cliConfig.RandomPort,
		DiscoveryService:  []DiscoveryService{dhtService},
		NetworkServices:   []NetworkService{vpnService},
		Logger:            logger,
		InterfaceAddress:  cliConfig.InterfaceAddress,
		InterfaceMTU:      cliConfig.InterfaceMTU,
		AdditionalOptions: []libp2p.Option{},
		Options:           libp2pOpts,
		DiscoveryPeers:    discoveryPeers,
		Sealer:            &crypto.DefaultSealer{},
	}

	return &Node{
		config: nodeConfig,
	}, nil
}

// Start joins the node over the p2p network
func (e *Node) Start(ctx context.Context) error {

	e.config.Logger.Info("Starting Solo P2P network")

	// Startup libp2p network
	err := e.startNetwork(ctx)
	if err != nil {
		return err
	}

	// Start eventual declared NetworkServices
	var networkServices sync.WaitGroup
	for _, s := range e.config.NetworkServices {
		err := s.Run(ctx, e.config.Logger, e.Host(), e.Broadcaster)
		if err != nil {
			return fmt.Errorf("error while starting network service: '%w'", err)
		}
		networkServices.Add(1)
	}

	// Wait for all Network Services to complete
	networkServices.Wait()

	return nil
}

func (e *Node) startNetwork(ctx context.Context) error {
	e.config.Logger.Debug("Generating host data")

	host, err := e.genHost(ctx)
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}
	e.host = host

	e.config.Logger.Info("Node ID:", host.ID())
	e.config.Logger.Info("Node Addresses:", host.Addrs())

	e.Broadcaster = broadcast.NewBroadcaster(
		e.config.Logger,
		&e.config.BroadcastKey,
		1024,
	)

	// Configure Broadcast and PRP
	myIP, _, err := net.ParseCIDR(e.config.InterfaceAddress)
	if err != nil {
		return err
	}
	go e.Broadcaster.Start(ctx, host, myIP.String())

	for _, sd := range e.config.DiscoveryService {
		if err := sd.Run(e.config.Logger, ctx, host); err != nil {
			e.config.Logger.Fatal(fmt.Errorf("while starting service discovery %+v: '%w", sd, err))
		}
	}

	e.config.Logger.Debug("Network started")
	return nil
}
