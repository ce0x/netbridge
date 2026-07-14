package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats [profile]",
	Short: "Show traffic statistics",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"stats","data":{"bytes_up":0,"bytes_down":0,"rate_up":0,"rate_down":0}}`)
			return nil
		}
		fmt.Println("No active session.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
