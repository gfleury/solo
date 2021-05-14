/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	configpackage "github.com/gfleury/solo/client/config"
	"github.com/gfleury/solo/client/node"
)

var (
	config = configpackage.Config{
		Token:             "",
		InterfaceAddress:  "10.1.0.1/24",
		InterfaceName:     "utun0",
		CreateInterface:   false,
		Libp2pLogLevel:    "",
		LogLevel:          "",
		DiscoveryPeers:    []string{},
		DiscoveryInterval: 0,
		InterfaceMTU:      0,
		MaxConnections:    0,
		HolePunch:         false,
		NatMap:            false,
		NatService:        false,
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

	go handleStopSignals()

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
	rootCmd.PersistentFlags().StringVarP(&config.InterfaceAddress, "address", "a", "10.1.0.1/24", "TUN interface ip address")
	rootCmd.PersistentFlags().StringVarP(&config.InterfaceName, "interface", "i", "utun0", "TUN interface name")
	rootCmd.PersistentFlags().BoolVarP(&config.CreateInterface, "create-iface", "c", true, "Create TUN network interface")
	rootCmd.PersistentFlags().StringVar(&config.Libp2pLogLevel, "libp2p-log-level", "error", "Libp2p log level")
	rootCmd.PersistentFlags().StringVarP(&config.LogLevel, "log-level", "l", "info", "Log level")
	rootCmd.PersistentFlags().StringArrayVarP(&config.DiscoveryPeers, "discovery-peers", "d", nil, "Discovery peers addresss")
	rootCmd.PersistentFlags().IntVarP(&config.DiscoveryInterval, "discovery-interval", "I", 600, "Discovery peers interval")
	rootCmd.PersistentFlags().IntVarP(&config.InterfaceMTU, "interface-mtu", "m", 1420, "Discovery peers interval")
	rootCmd.PersistentFlags().IntVarP(&config.MaxConnections, "max-connections", "M", 256, "Maximum peer connections")
	rootCmd.PersistentFlags().BoolVarP(&config.HolePunch, "hole-punch", "H", false, "Enable holepunch to bypass NAT")
	rootCmd.PersistentFlags().BoolVarP(&config.NatMap, "nat-map", "N", false, "Enable NAT map")
	rootCmd.PersistentFlags().BoolVarP(&config.NatService, "nat-service", "n", false, "Enable NAT service")

	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
