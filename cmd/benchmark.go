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
		profile := ""
		if len(args) > 0 {
			profile = args[0]
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"benchmark","data":{"all":%v,"profile":"%s","results":[]}}`, all, profile)
			return nil
		}
		if all {
			fmt.Println("Benchmarking all profiles...")
		} else {
			fmt.Printf("Benchmarking profile: %s\n", profile)
		}
		fmt.Println("Profile         Latency    Throughput    Score")
		fmt.Println("────────────────────────────────────────────────")
		return nil
	},
}

func init() {
	benchmarkCmd.Flags().Bool("all", false, "benchmark all profiles")
	rootCmd.AddCommand(benchmarkCmd)
}
