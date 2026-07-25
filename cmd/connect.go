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

		eng, err := getEngine()
		if err != nil {
			return err
		}

		mgr := eng.ProfileManager()
		ctx := cmd.Context()

		var profileID string
		if len(args) > 0 {
			p, err := resolveProfile(ctx, mgr, args[0])
			if err != nil {
				if jsonOutput {
					fmt.Printf(`{"success":false,"error":"%s"}`, err)
					return nil
				}
				return err
			}
			profileID = p.ID
		} else {
			active, err := mgr.GetActive(ctx)
			if err != nil {
				if jsonOutput {
					fmt.Printf(`{"success":false,"error":"no active profile: %s"}`, err)
					return nil
				}
				return fmt.Errorf("no active profile: %w", err)
			}
			profileID = active.ID
		}

		if port > 0 {
			mode = string(netbridge.ModeSOCKS)
		}

		sess, err := eng.SessionManager().Connect(ctx, profileID, netbridge.SessionMode(mode))
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		// Set as active profile for health/test/status commands
		_ = mgr.SetActive(ctx, profileID)

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"connect","data":{"session_id":"%s","profile_id":"%s","mode":"%s","local_addr":"%s"}}`,
				sess.ID, sess.ProfileID, sess.Mode, sess.LocalAddr)
			return nil
		}

		fmt.Printf("Connected: session %s (profile %s, mode %s, local %s)\n",
			sess.ID, sess.ProfileID, sess.Mode, sess.LocalAddr)
		return nil
	},
}

func init() {
	connectCmd.Flags().String("mode", "socks", "connection mode: socks|http|tun")
	connectCmd.Flags().Int("port", 0, "local port override")
	connectCmd.Flags().Bool("no-watchdog", false, "disable auto-reconnect for this session")
	rootCmd.AddCommand(connectCmd)
}
