/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/gfleury/solo/client/node"
	"github.com/spf13/cobra"
)

// generateTokenCmd represents the generateToken command
var generateTokenCmd = &cobra.Command{
	Use:   "generateToken",
	Short: "Generate a VPN token to use",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		newConfig := node.GenerateNewConnectionData()
		fmt.Println(newConfig.Base64())
	},
}

func init() {
	rootCmd.AddCommand(generateTokenCmd)
}
