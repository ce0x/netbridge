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
		eng, err := getEngine()
		if err != nil {
			return err
		}

		presets := eng.DNSEngine().ListPresets()

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_list","data":{"presets":[`)
			for i, p := range presets {
				if i > 0 {
					fmt.Print(",")
				}
				fmt.Printf(`{"name":"%s","servers":%v}`, p.Name, p.Servers)
			}
			fmt.Print(`]}}`)
			return nil
		}

		fmt.Println("DNS Presets:")
		for _, p := range presets {
			fmt.Printf("  %-12s — %v\n", p.Name, p.Servers)
		}
		return nil
	},
}

var dnsUseCmd = &cobra.Command{
	Use:   "use <preset|ip>",
	Short: "Set active DNS resolver",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if err := eng.DNSEngine().SetResolver(cmd.Context(), args[0]); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

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
		eng, err := getEngine()
		if err != nil {
			return err
		}

		results, err := eng.DNSEngine().Benchmark(cmd.Context())
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_benchmark","data":{"results":[`)
			for i, r := range results {
				if i > 0 {
					fmt.Print(",")
				}
				fmt.Printf(`{"name":"%s","server":"%s","latency_ms":%d,"error":"%s"}`,
					r.Name, r.Server, r.Latency.Milliseconds(), errStr(r.Error))
			}
			fmt.Print(`]}}`)
			return nil
		}

		fmt.Printf("%-12s %-20s %s\n", "Name", "Server", "Latency")
		fmt.Println("─────────────────────────────────────────")
		for _, r := range results {
			if r.Error != nil {
				fmt.Printf("%-12s %-20s ERROR: %s\n", r.Name, r.Server, r.Error)
			} else {
				fmt.Printf("%-12s %-20s %s\n", r.Name, r.Server, r.Latency.Round(1e6))
			}
		}
		return nil
	},
}

var dnsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current DNS configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		current := eng.DNSEngine().CurrentResolver()

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"dns_show","data":{"current":"%s"}}`, current)
			return nil
		}
		fmt.Printf("Current DNS: %s\n", current)
		return nil
	},
}

var dnsResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Restore system default DNS",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if err := eng.DNSEngine().Reset(cmd.Context()); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"dns_reset","data":{"status":"reset"}}`)
			return nil
		}
		fmt.Println("DNS restored to system default.")
		return nil
	},
}

func errStr(e error) string {
	if e == nil {
		return ""
	}
	return e.Error()
}

func init() {
	dnsCmd.AddCommand(dnsListCmd, dnsUseCmd, dnsBenchCmd, dnsShowCmd, dnsResetCmd)
	rootCmd.AddCommand(dnsCmd)
}
