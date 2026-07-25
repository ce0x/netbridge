package main

import (
	"fmt"

	"github.com/spf13/cobra"
	netbridge "github.com/netbridge/netbridge"
)

var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "Smart routing engine commands",
}

var routeAddCmd = &cobra.Command{
	Use:   "add <domain|pattern> <profile>",
	Short: "Add a routing rule",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		mgr := eng.ProfileManager()
		ctx := cmd.Context()

		p, err := resolveProfile(ctx, mgr, args[1])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		rule := netbridge.RouteRule{
			ID:        fmt.Sprintf("rule-%d", len(args)),
			Pattern:   args[0],
			RuleType:  "domain",
			ProfileID: p.ID,
			Priority:  100,
			Enabled:   true,
		}

		if err := eng.RoutingEngine().AddRule(ctx, rule); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"route_add","data":{"pattern":"%s","profile":"%s"}}`, args[0], p.ID)
			return nil
		}
		fmt.Printf("Route added: %s → %s\n", args[0], p.Name)
		return nil
	},
}

var routeRemoveCmd = &cobra.Command{
	Use:   "remove <domain|pattern>",
	Short: "Remove a routing rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if err := eng.RoutingEngine().RemoveRule(cmd.Context(), args[0]); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"route_remove","data":{"pattern":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Route removed: %s\n", args[0])
		return nil
	},
}

var routeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List routing rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		rules, err := eng.RoutingEngine().ListRules(cmd.Context())
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"route_list","data":{"rules":[`)
			for i, r := range rules {
				if i > 0 {
					fmt.Print(",")
				}
				fmt.Printf(`{"id":"%s","pattern":"%s","profile_id":"%s","enabled":%v}`,
					r.ID, r.Pattern, r.ProfileID, r.Enabled)
			}
			fmt.Print(`]}}`)
			return nil
		}

		if len(rules) == 0 {
			fmt.Println("No routing rules configured.")
			return nil
		}
		for _, r := range rules {
			status := "✓"
			if !r.Enabled {
				status = "✗"
			}
			fmt.Printf("  %s %s → %s\n", status, r.Pattern, r.ProfileID)
		}
		return nil
	},
}

var routeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all routing rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if err := eng.RoutingEngine().ClearRules(cmd.Context()); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"route_clear","data":{}}`)
			return nil
		}
		fmt.Println("All routing rules cleared.")
		return nil
	},
}

func init() {
	routeCmd.AddCommand(routeAddCmd, routeRemoveCmd, routeListCmd, routeClearCmd)
	rootCmd.AddCommand(routeCmd)
}
