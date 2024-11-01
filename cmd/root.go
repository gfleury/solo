/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/spf13/cobra"

	configpackage "github.com/gfleury/solo/client/config"
	"github.com/gfleury/solo/client/node"
)

var (
	DEFAULT_DISCOVERY_PEERS = []string{"/dnsaddr/solo-rendezvous.fleury.gg/p2p/12D3KooWGXAXwKmP4Pg3QWUnrghQaJiHLJrKScSVpTUn59hGT7Vh"}
)

var (
	config = configpackage.Config{
		Token:                "",
		InterfaceAddress:     "10.1.0.1/24",
		InterfaceName:        "utun0",
		CreateInterface:      false,
		Libp2pLogLevel:       "",
		LogLevel:             "",
		DiscoveryPeers:       []string{},
		PublicDiscoveryPeers: false,
		DiscoveryInterval:    0,
		InterfaceMTU:         0,
		MaxConnections:       0,
		HolePunch:            false,
		StandaloneMode:       false,
		RandomIdentity:       false,
		RandomPort:           false,
		PublishLocalRoutes:   false,
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "solo",
	Short: "Solo P2P standalone VPN service",
	Long:  "Full descentralized P2P VPN service",
	Run:   runMain,
}

func runMain(cmd *cobra.Command, args []string) {

	e, err := node.NewWithConfig(config)
	if err != nil {
		fmt.Printf("failed to create new node: %s\n", err)
		return
	}

	ctx := context.Background()

	go handleStopSignals(e)

	err = e.Start(ctx)
	if err != nil {
		fmt.Printf("failed to start node: %s\n", err)
		return
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().StringVarP(&config.Token, "token", "t", "", "Configuration token")
	rootCmd.PersistentFlags().StringVarP(&config.InterfaceAddress, "address", "a", "192.168.254.0/24", "TUN interface ip address")
	rootCmd.PersistentFlags().StringVarP(&config.InterfaceName, "interface", "i", "utun0", "TUN interface name")
	rootCmd.PersistentFlags().BoolVarP(&config.CreateInterface, "create-iface", "c", true, "Create TUN network interface")
	rootCmd.PersistentFlags().BoolVarP(&config.PublishLocalRoutes, "publish-local-routes", "r", false, "Publish local routes to other hosts")
	rootCmd.PersistentFlags().StringVar(&config.Libp2pLogLevel, "libp2p-log-level", "error", "Libp2p log level")
	rootCmd.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "l", "info", "Log level")
	rootCmd.PersistentFlags().StringArrayVarP(&config.DiscoveryPeers, "discovery-peers", "d", DEFAULT_DISCOVERY_PEERS, "Discovery peers addresss")
	rootCmd.PersistentFlags().IntVarP(&config.DiscoveryInterval, "discovery-interval", "I", 10, "Discovery peers interval")
	rootCmd.PersistentFlags().IntVarP(&config.InterfaceMTU, "interface-mtu", "m", 1412, "Discovery peers interval")
	rootCmd.PersistentFlags().IntVarP(&config.MaxConnections, "max-connections", "M", 256, "Maximum peer connections")
	rootCmd.PersistentFlags().BoolVarP(&config.HolePunch, "hole-punch", "H", true, "Enable holepunch to bypass NAT")
	rootCmd.PersistentFlags().BoolVarP(&config.PublicDiscoveryPeers, "public", "p", false, "Enable public discovery peers")
	rootCmd.PersistentFlags().BoolVarP(&config.StandaloneMode, "standalone", "s", false, "Enable standalone mode")

	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	go http.ListenAndServe(":7777", mux)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
