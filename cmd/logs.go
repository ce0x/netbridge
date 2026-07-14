package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View application logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		follow, _ := cmd.Flags().GetBool("follow")
		level, _ := cmd.Flags().GetString("level")
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"logs","data":{"follow":%v,"level":"%s"}}`, follow, level)
			return nil
		}
		fmt.Println("No log entries.")
		return nil
	},
}

func init() {
	logsCmd.Flags().BoolP("follow", "f", false, "follow log output")
	logsCmd.Flags().String("level", "info", "filter by level: debug|info|warn|error")
	rootCmd.AddCommand(logsCmd)
}
