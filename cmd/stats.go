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
		eng, err := getEngine()
		if err != nil {
			return err
		}

		stats := eng.SessionManager().Stats()

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"stats","data":{"bytes_up":%d,"bytes_down":%d,"rate_up":0,"rate_down":0,"uptime_sec":%.0f}}`,
				stats.BytesUp, stats.BytesDown, stats.Uptime.Seconds())
			return nil
		}

		fmt.Printf("Bytes up:   %d\n", stats.BytesUp)
		fmt.Printf("Bytes down: %d\n", stats.BytesDown)
		fmt.Printf("Uptime:     %s\n", stats.Uptime.Round(1e6))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
