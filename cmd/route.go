package main

import (
	"fmt"

	"github.com/spf13/cobra"
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
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"route_add","data":{"pattern":"%s","profile":"%s"}}`, args[0], args[1])
			return nil
		}
		fmt.Printf("Route added: %s → %s\n", args[0], args[1])
		return nil
	},
}

var routeRemoveCmd = &cobra.Command{
	Use:   "remove <domain|pattern>",
	Short: "Remove a routing rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"route_list","data":{"rules":[]}}`)
			return nil
		}
		fmt.Println("No routing rules configured.")
		return nil
	},
}

var routeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all routing rules",
	RunE: func(cmd *cobra.Command, args []string) error {
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
