package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <command...>",
	Short: "Run command through active profile",
	Long:  `Executes the specified command with proxy environment variables set.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"run","data":{"argv":%v}}`, args)
			return nil
		}

		return eng.RunCommand(cmd.Context(), "", args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
