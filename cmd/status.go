package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"status","data":{"status":"disconnected","profile":null,"backend":null,"mode":null}}`)
			return nil
		}
		fmt.Println("● Disconnected")
		fmt.Println("  No active session.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
