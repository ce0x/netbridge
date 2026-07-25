package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test [profile]",
	Short: "Test profile connectivity",
	Long: `Verifies DNS resolution, TCP reachability, TLS handshake,
latency, download throughput, and upload throughput.`,
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
			fmt.Printf(`{"success":true,"command":"test","data":{"profile_id":"%s","reachable":%v,"latency_ms":%d,"packet_loss":%.2f,"error":"%s"}}`,
				result.ProfileID, result.Reachable, result.Latency.Milliseconds(), result.PacketLoss, result.Error)
			return nil
		}

		fmt.Printf("Testing profile: %s\n", result.ProfileID)
		if result.Reachable {
			fmt.Println("  Reachable:    ✓")
		} else {
			fmt.Println("  Reachable:    ✗")
		}
		fmt.Printf("  Latency:      %s\n", result.Latency.Round(1e6))
		fmt.Printf("  Packet loss:  %.1f%%\n", result.PacketLoss*100)
		if result.Error != "" {
			fmt.Printf("  Error:        %s\n", result.Error)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
}
