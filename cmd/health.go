package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health [profile]",
	Short: "Health check for profile connectivity",
	Long: `Reports reachability, latency (avg, min, max),
packet loss percentage, and protocol-level verification.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
					fmt.Printf(`{"success":false,"error":"no active profile"}`)
					return nil
				}
				return fmt.Errorf("no active profile: %w", err)
			}
			profileID = active.ID
		}

		result, err := eng.HealthEngine().Check(ctx, profileID)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"health","data":{"profile_id":"%s","reachable":%v,"latency_ms":%d,"packet_loss":%.2f}}`,
				result.ProfileID, result.Reachable, result.Latency.Milliseconds(), result.PacketLoss)
			return nil
		}

		fmt.Printf("Health check for: %s\n", result.ProfileID)
		fmt.Printf("  Reachable:   %v\n", result.Reachable)
		fmt.Printf("  Latency:     %s\n", result.Latency.Round(1e6))
		fmt.Printf("  Packet loss: %.1f%%\n", result.PacketLoss*100)
		if result.Error != "" {
			fmt.Printf("  Error:       %s\n", result.Error)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
