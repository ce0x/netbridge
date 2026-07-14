package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect current session",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"disconnect","data":{"status":"disconnected"}}`)
			return nil
		}
		fmt.Println("Disconnected.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
