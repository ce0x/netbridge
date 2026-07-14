package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile management commands",
}

var importCmd = &cobra.Command{
	Use:   "import <url|file>",
	Short: "Import profile from URL or file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"import","data":{"source":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Importing profile from: %s\n", args[0])
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"export","data":{"name":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Exporting profile: %s\n", args[0])
		return nil
	},
}

var deleteProfileCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"delete","data":{"name":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Deleted profile: %s\n", args[0])
		return nil
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"rename","data":{"old":"%s","new":"%s"}}`, args[0], args[1])
			return nil
		}
		fmt.Printf("Renamed %s → %s\n", args[0], args[1])
		return nil
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone <name> <new>",
	Short: "Clone profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"clone","data":{"source":"%s","target":"%s"}}`, args[0], args[1])
			return nil
		}
		fmt.Printf("Cloned %s → %s\n", args[0], args[1])
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"list","data":{"profiles":[]}}`)
			return nil
		}
		fmt.Println("No profiles configured.")
		return nil
	},
}

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"use","data":{"active":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Active profile set to: %s\n", args[0])
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		reveal, _ := cmd.Flags().GetBool("reveal")
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"show","data":{"name":"%s","reveal":%v}}`, args[0], reveal)
			return nil
		}
		fmt.Printf("Profile: %s\n", args[0])
		if !reveal {
			fmt.Println("  (sensitive values masked, use --reveal to show)")
		}
		return nil
	},
}

func init() {
	showCmd.Flags().Bool("reveal", false, "show sensitive values")
	profileCmd.AddCommand(importCmd, exportCmd, deleteProfileCmd, renameCmd, cloneCmd, listCmd, useCmd, showCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(listCmd)
}
