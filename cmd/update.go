package main

import (
	"fmt"

	"github.com/netbridge/netbridge/internal/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update netbridge to latest version",
}

var updateCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if a new version is available",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := selfupdate.CheckLatest(cmd.Context())
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"update_check","data":{"current":"%s","latest":"%s","update_available":%v}}`,
				info.Current, info.Latest, info.UpdateAvailable)
			return nil
		}

		if info.UpdateAvailable {
			fmt.Printf("Update available: %s → %s\n", info.Current, info.Latest)
			fmt.Println("Run 'netbridge update install' to update.")
		} else {
			fmt.Printf("Already up to date (%s)\n", info.Current)
		}
		return nil
	},
}

var updateInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Download and install the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := selfupdate.CheckLatest(cmd.Context())
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if !info.UpdateAvailable {
			if jsonOutput {
				fmt.Printf(`{"success":true,"command":"update_install","data":{"status":"up_to_date","version":"%s"}}`, info.Current)
				return nil
			}
			fmt.Printf("Already up to date (%s)\n", info.Current)
			return nil
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"update_install","data":{"status":"downloading","from":"%s","to":"%s"}}`,
				info.Current, info.Latest)
			return nil
		}

		fmt.Printf("Updating %s → %s...\n", info.Current, info.Latest)
		if err := selfupdate.Install(cmd.Context(), info); err != nil {
			return fmt.Errorf("install update: %w", err)
		}

		fmt.Printf("Updated to %s. Restart to apply.\n", info.Latest)
		return nil
	},
}

func init() {
	updateCmd.AddCommand(updateCheckCmd, updateInstallCmd)
	rootCmd.AddCommand(updateCmd)
}
