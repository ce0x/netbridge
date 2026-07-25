package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect current session",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if err := eng.SessionManager().Disconnect(cmd.Context()); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"disconnect","data":{"status":"disconnected"}}`)
			return nil
		}
		fmt.Println("Disconnected.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
