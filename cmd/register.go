/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/gfleury/solo/client/node"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the node on the server",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		node, err := node.NewWithConfig(config)
		if err != nil {
			fmt.Printf("failed to create new node: %s\n", err)
			return
		}
		err = node.Register(context.Background())
		if err != nil {
			fmt.Printf("failed to register node: %s\n", err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
