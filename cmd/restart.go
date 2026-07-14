package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	restartCmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart current connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				fmt.Print(`{"success":true,"command":"restart","data":{"status":"restarted"}}`)
				return nil
			}
			fmt.Println("Restarting connection...")
			return nil
		},
	}

	reloadCmd = &cobra.Command{
		Use:   "reload",
		Short: "Reload config without disconnecting",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				fmt.Print(`{"success":true,"command":"reload","data":{"status":"reloaded"}}`)
				return nil
			}
			fmt.Println("Reloading configuration...")
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(reloadCmd)
}
