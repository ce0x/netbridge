package main

import (
	"fmt"

	"github.com/spf13/cobra"
	netbridge "github.com/netbridge/netbridge"
)

var connectCmd = &cobra.Command{
	Use:   "connect [profile]",
	Short: "Connect active or named profile",
	Long: `Connects the selected profile. If no profile is specified,
uses the currently active profile.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, _ := cmd.Flags().GetString("mode")
		port, _ := cmd.Flags().GetInt("port")

		if mode == "" {
			mode = string(netbridge.ModeSOCKS)
		}

		var profileName string
		if len(args) > 0 {
			profileName = args[0]
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"connect","data":{"profile":"%s","mode":"%s","port":%d}}`, profileName, mode, port)
			return nil
		}

		fmt.Printf("Connecting profile %q in %s mode...\n", profileName, mode)
		return nil
	},
}

func init() {
	connectCmd.Flags().String("mode", "socks", "connection mode: socks|http|tun")
	connectCmd.Flags().Int("port", 0, "local port override")
	connectCmd.Flags().Bool("no-watchdog", false, "disable auto-reconnect for this session")
	rootCmd.AddCommand(connectCmd)
}
