package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark [--all] [profile]",
	Short: "Benchmark and score profile performance",
	Long: `Measures latency, throughput, jitter, and stability over time.
Outputs a scored comparison table.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		eng, err := getEngine()
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		be := eng.BenchmarkEngine()

		if all {
			results, err := be.RunAll(ctx)
			if err != nil {
				if jsonOutput {
					fmt.Printf(`{"success":false,"error":"%s"}`, err)
					return nil
				}
				return err
			}

			if jsonOutput {
				fmt.Print(`{"success":true,"command":"benchmark","data":{"results":[`)
				for i, r := range results {
					if i > 0 {
						fmt.Print(",")
					}
					fmt.Printf(`{"profile_id":"%s","latency_ms":%d,"jitter_ms":%d,"throughput_bps":%.0f,"packet_loss":%.2f,"score":%d}`,
						r.ProfileID, r.Latency.Milliseconds(), r.Jitter.Milliseconds(), r.Throughput, r.PacketLoss, r.Score)
				}
				fmt.Print(`]}}`)
				return nil
			}

			fmt.Printf("%-20s %-12s %-12s %-14s %-6s %s\n", "Profile", "Latency", "Jitter", "Throughput", "Loss", "Score")
			fmt.Println("────────────────────────────────────────────────────────────────────")
			for _, r := range results {
				fmt.Printf("%-20s %-12s %-12s %-14s %-5.1f%% %d\n",
					r.ProfileID, r.Latency.Round(1e6), r.Jitter.Round(1e6),
					fmt.Sprintf("%.0f bps", r.Throughput), r.PacketLoss*100, r.Score)
			}
			return nil
		}

		if len(args) == 0 {
			active, err := eng.ProfileManager().GetActive(ctx)
			if err != nil {
				if jsonOutput {
					fmt.Printf(`{"success":false,"error":"no active profile"}`)
					return nil
				}
				return fmt.Errorf("no active profile: %w", err)
			}
			args = []string{active.ID}
		}

		p, err := resolveProfile(ctx, eng.ProfileManager(), args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		result, err := be.Run(ctx, p.ID)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"benchmark","data":{"profile_id":"%s","latency_ms":%d,"jitter_ms":%d,"throughput_bps":%.0f,"packet_loss":%.2f,"score":%d}}`,
				result.ProfileID, result.Latency.Milliseconds(), result.Jitter.Milliseconds(), result.Throughput, result.PacketLoss, result.Score)
			return nil
		}

		fmt.Printf("Benchmark: %s\n", result.ProfileID)
		fmt.Printf("  Latency:    %s\n", result.Latency.Round(1e6))
		fmt.Printf("  Jitter:     %s\n", result.Jitter.Round(1e6))
		fmt.Printf("  Throughput: %.0f bps\n", result.Throughput)
		fmt.Printf("  Packet loss: %.1f%%\n", result.PacketLoss*100)
		fmt.Printf("  Score:      %d\n", result.Score)
		return nil
	},
}

func init() {
	benchmarkCmd.Flags().Bool("all", false, "benchmark all profiles")
	rootCmd.AddCommand(benchmarkCmd)
}
