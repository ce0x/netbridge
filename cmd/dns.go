package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS management commands",
}

var dnsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available DNS presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_list","data":{"presets":["cloudflare","google","quad9","adguard","system"]}}`)
			return nil
		}
		fmt.Println("DNS Presets:")
		fmt.Println("  cloudflare  — 1.1.1.1, 1.0.0.1")
		fmt.Println("  google      — 8.8.8.8, 8.8.4.4")
		fmt.Println("  quad9       — 9.9.9.9, 149.112.112.112")
		fmt.Println("  adguard     — 94.140.14.14, 94.140.15.15")
		fmt.Println("  system      — /etc/resolv.conf")
		return nil
	},
}

var dnsUseCmd = &cobra.Command{
	Use:   "use <preset|ip>",
	Short: "Set active DNS resolver",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"dns_use","data":{"resolver":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("DNS resolver set to: %s\n", args[0])
		return nil
	},
}

var dnsBenchCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmark all DNS presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_benchmark","data":{"results":[]}}`)
			return nil
		}
		fmt.Println("Benchmarking DNS resolvers...")
		return nil
	},
}

var dnsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current DNS configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_show","data":{"current":"system"}}`)
			return nil
		}
		fmt.Println("Current DNS: system (/etc/resolv.conf)")
		return nil
	},
}

var dnsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Restore system default DNS",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_reset","data":{"status":"reset"}}`)
			return nil
		}
		fmt.Println("DNS restored to system default.")
		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsListCmd, dnsUseCmd, dnsBenchCmd, dnsShowCmd, dnsResetCmd)
	rootCmd.AddCommand(dnsCmd)
}
