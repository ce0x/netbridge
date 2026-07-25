package main

import (
	"fmt"

	"github.com/netbridge/netbridge/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI dashboard",
	Long:  `Launches the full-screen terminal dashboard with real-time status, profile management, and quick actions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"tui","data":{"status":"launching"}}`)
			return nil
		}

		eng, err := getEngine()
		if err != nil {
			return err
		}

		return tui.RunApp(eng)
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
