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
	"net/http"
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
	"github.com/gfleury/solo/common"
	"github.com/gfleury/solo/common/models"
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
	// dhtService.OTPKey = connectionCfg.DiscoveryKey
	dhtService.DiscoveryPeers = discoveryPeers

	// Configure VPN
	vpnService := vpn.VPNNetworkService(vpn.InterfaceConfig{
		InterfaceMTU:     cliConfig.InterfaceMTU,
		InterfaceName:    cliConfig.InterfaceName,
		InterfaceAddress: cliConfig.InterfaceAddress,
		CreateInterface:  cliConfig.CreateInterface,
		// PreSharedKey:     connectionCfg.VPNPreSharedKey,
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

	nodeConfig := Config{
		// BroadcastKey:      connectionCfg.BroadcastKey,
		ListenAddresses:       []discovery.AddrList{},
		RandomIdentity:        cliConfig.RandomIdentity,
		RandomPort:            cliConfig.RandomPort,
		DiscoveryService:      []DiscoveryService{dhtService},
		NetworkServices:       []NetworkService{vpnService},
		Logger:                logger,
		InterfaceAddress:      cliConfig.InterfaceAddress,
		InterfaceMTU:          cliConfig.InterfaceMTU,
		AdditionalOptions:     []libp2p.Option{},
		Options:               libp2pOpts,
		DiscoveryPeers:        discoveryPeers,
		Sealer:                &crypto.DefaultSealer{},
		PublicDiscoveryPeers:  cliConfig.PublicDiscoveryPeers,
		ConnectionConfigToken: cliConfig.Token,
	}

	return &Node{
		config: nodeConfig,
	}, nil
}

func (e *Node) Register(ctx context.Context) error {
	var err error

	// Startup libp2p network
	e.host, err = e.genHost(ctx)
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}

	e.config.Logger.Info("Node ID:", e.host.ID())
	e.config.Logger.Info("Node Addresses:", e.host.Addrs())

	peerInfo, err := peer.AddrInfoFromP2pAddr(e.config.DiscoveryPeers[0])
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}

	err = e.host.Connect(ctx, *peerInfo)
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}

	client := common.GetSoloAPIP2PClient(peerInfo.ID, e.host)

	code, err := client.RegisterNode(models.NewLocalNode(e.host, ""))
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}

	fmt.Println("Go to the web interface and enter the code", code)

	return nil
}

func (e *Node) configurationDiscovery(ctx context.Context) error {
	var connectionCfg *models.YAMLConnectionConfig
	var err error

	if e.config.PublicDiscoveryPeers {
		connectionCfg, err = models.YAMLConnectionConfigFromToken(e.config.ConnectionConfigToken)
		if err != nil {
			return err
		}
	} else {
	OUT:
		for {
			for _, peerID := range e.host.Peerstore().PeersWithKeys() {
				if peerID == e.host.ID() {
					// Skip ourselves
					continue
				}
				client := common.GetSoloAPIP2PClient(peerID, e.host)

				cfg, statusCode, err := client.GetNodeNetworkConfiguration()
				if err != nil {
					switch statusCode {
					case http.StatusNotFound:
						return fmt.Errorf("node not found, register the node first: %s", err)
					case http.StatusFailedDependency:
						e.config.Logger.Errorf("node is not activated yet, go to interface and enter code")
						time.Sleep(10 * time.Second)
						continue
					default:
						e.config.Logger.Errorf("failed to discovery configuration from: %s with %s", peerID, err)
						time.Sleep(10 * time.Second)
						continue
					}
				}
				e.config.InterfaceAddress = cfg.InterfaceAddress
				connectionCfg, err = models.YAMLConnectionConfigFromToken(cfg.ConnectionConfigToken)
				if err != nil {
					return err
				}
				break OUT
			}
		}

	}

	// Fill last configuration items from Connection Token
	e.config.DiscoveryService[0].(*discovery.DHT).OTPKeyReceiver <- connectionCfg.DiscoveryKey
	e.config.NetworkServices[0].(*vpn.VPNService).Config.PreSharedKey = connectionCfg.VPNPreSharedKey
	e.config.BroadcastKey = connectionCfg.BroadcastKey

	return nil
}

// Start joins the node over the p2p network
func (e *Node) Start(ctx context.Context) error {
	var err error

	e.config.Logger.Info("Starting Solo P2P network")

	// Startup libp2p network
	e.host, err = e.genHost(ctx)
	if err != nil {
		e.config.Logger.Error(err.Error())
		return err
	}

	e.config.Logger.Info("Node ID:", e.host.ID())
	e.config.Logger.Info("Node Addresses:", e.host.Addrs())

	// Startup discovery
	err = e.startDiscovery(ctx)
	if err != nil {
		return err
	}

	err = e.configurationDiscovery(ctx)
	if err != nil {
		return err
	}

	// Startup Broadcaster
	err = e.startBroadcastService(ctx)
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

func (e *Node) startDiscovery(ctx context.Context) error {
	for _, sd := range e.config.DiscoveryService {
		if err := sd.Run(e.config.Logger, ctx, e.host); err != nil {
			e.config.Logger.Fatal(fmt.Errorf("while starting service discovery %+v: '%w", sd, err))
		}
	}

	e.config.Logger.Debug("Network started")
	return nil
}

func (e *Node) startBroadcastService(ctx context.Context) error {
	e.Broadcaster = broadcast.NewStreamBroadcaster(
		e.config.Logger,
		e.config.DiscoveryPeers,
		e.config.BroadcastKey,
	)

	// Configure Broadcast and PRP
	myIP, _, err := net.ParseCIDR(e.config.InterfaceAddress)
	if err != nil {
		return err
	}
	go e.Broadcaster.Start(ctx, e.host, myIP.String())

	return nil
}
