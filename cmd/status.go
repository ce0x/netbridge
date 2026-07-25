package main

import (
	"fmt"

	"github.com/spf13/cobra"
	netbridge "github.com/netbridge/netbridge"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		sm := eng.SessionManager()
		status := sm.Status()

		if status == netbridge.StatusDisconnected {
			if jsonOutput {
				fmt.Print(`{"success":true,"command":"status","data":{"status":"disconnected","profile":null,"session":null}}`)
				return nil
			}
			fmt.Println("● Disconnected")
			fmt.Println("  No active session.")
			return nil
		}

		sess, err := sm.Current()
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"status","data":{"status":"%s","session_id":"%s","profile_id":"%s","mode":"%s","local_addr":"%s","uptime_sec":%.0f}}`,
				status, sess.ID, sess.ProfileID, sess.Mode, sess.LocalAddr, sess.StartedAt.Sub(sess.StartedAt).Seconds())
			return nil
		}

		fmt.Printf("● %s\n", status)
		fmt.Printf("  Session:   %s\n", sess.ID)
		fmt.Printf("  Profile:   %s\n", sess.ProfileID)
		fmt.Printf("  Mode:      %s\n", sess.Mode)
		fmt.Printf("  Local:     %s\n", sess.LocalAddr)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
