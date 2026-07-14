package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Live connection dashboard",
	Long: `Displays active profile and backend, uptime,
current upload/download rate, total traffic, and active connections.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"top","data":{"active":false}}`)
			return nil
		}
		fmt.Println("No active session. Use 'netbridge connect' to start.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(topCmd)
}
