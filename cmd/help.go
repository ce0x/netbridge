package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help [command]",
	Short: "Show help for any command",
	Long:  `Detailed help with description, options, examples, and related commands.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return rootCmd.Help()
		}
		found, _, err := rootCmd.Find(args)
		if err != nil {
			return fmt.Errorf("unknown command: %s", args[0])
		}
		return found.Help()
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}
