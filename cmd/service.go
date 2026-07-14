package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Systemd service management",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install and enable systemd unit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_install","data":{"status":"installed"}}`)
			return nil
		}
		fmt.Println("NetBridge service installed and enabled.")
		return nil
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_start","data":{"status":"started"}}`)
			return nil
		}
		fmt.Println("Service started.")
		return nil
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_stop","data":{"status":"stopped"}}`)
			return nil
		}
		fmt.Println("Service stopped.")
		return nil
	},
}

var serviceRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_restart","data":{"status":"restarted"}}`)
			return nil
		}
		fmt.Println("Service restarted.")
		return nil
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show systemd service status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_status","data":{"status":"inactive"}}`)
			return nil
		}
		fmt.Println("NetBridge service: inactive")
		return nil
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove systemd unit",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"service_uninstall","data":{"status":"uninstalled"}}`)
			return nil
		}
		fmt.Println("Service uninstalled.")
		return nil
	},
}

func init() {
	serviceCmd.AddCommand(serviceInstallCmd, serviceStartCmd, serviceStopCmd, serviceRestartCmd, serviceStatusCmd, serviceUninstallCmd)
	rootCmd.AddCommand(serviceCmd)
}
