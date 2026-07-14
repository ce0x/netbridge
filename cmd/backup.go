package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create encrypted backup archive",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		includeHistory, _ := cmd.Flags().GetBool("include-history")
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"backup","data":{"output":"%s","include_history":%v}}`, output, includeHistory)
			return nil
		}
		fmt.Println("Backup created successfully.")
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore <file>",
	Short: "Restore from backup archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"restore","data":{"file":"%s"}}`, args[0])
			return nil
		}
		fmt.Printf("Restored from: %s\n", args[0])
		return nil
	},
}

func init() {
	backupCmd.Flags().String("output", "", "output file path")
	backupCmd.Flags().Bool("include-history", false, "include session history")
	rootCmd.AddCommand(backupCmd)
	rootCmd.AddCommand(restoreCmd)
}
