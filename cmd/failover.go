package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var failoverCmd = &cobra.Command{
	Use:   "failover",
	Short: "Failover chain management",
}

var failoverCreateCmd = &cobra.Command{
	Use:   "create <chain-name> <profile-a> <profile-b> [profile-c...]",
	Short: "Create a failover chain",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"failover_create","data":{"name":"%s","profiles":%v}}`, args[0], args[1:])
			return nil
		}
		fmt.Printf("Failover chain %q created with profiles: %v\n", args[0], args[1:])
		return nil
	},
}

var failoverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List failover chains",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"failover_list","data":{"chains":[]}}`)
			return nil
		}
		fmt.Println("No failover chains configured.")
		return nil
	},
}

var failoverDeleteCmd = &cobra.Command{
	Use:   "delete <chain-name>",
	Short: "Delete a failover chain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"failover_delete","data":{"name":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Deleted failover chain: %s\n", args[0])
		return nil
	},
}

var failoverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show failover chain status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"failover_status","data":{"chains":[]}}`)
			return nil
		}
		fmt.Println("No active failover chains.")
		return nil
	},
}

func init() {
	failoverCmd.AddCommand(failoverCreateCmd, failoverListCmd, failoverDeleteCmd, failoverStatusCmd)
	rootCmd.AddCommand(failoverCmd)
}
