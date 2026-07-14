package main

import (
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	quietMode  bool
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "netbridge",
	Short: "NetBridge — universal network access and connectivity toolkit",
	Long:  `NetBridge is a unified connectivity layer between Linux applications and external networks.`,
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
}
