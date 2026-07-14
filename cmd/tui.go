package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI mode",
	Long:  `Launches the full-screen terminal user interface for visual menu-driven operation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"tui","data":{"status":"launching"}}`)
			return nil
		}
		fmt.Println("TUI mode — interactive terminal UI")
		fmt.Println("(requires terminal with color support)")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
