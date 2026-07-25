package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print export commands for proxy environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}

		envVars := eng.EnvVars()

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"env","data":{`)
			first := true
			for k, v := range envVars {
				if !first {
					fmt.Print(",")
				}
				fmt.Printf(`"%s":"%s"`, k, v)
				first = false
			}
			fmt.Print(`}}`)
			return nil
		}

		keys := []string{"http_proxy", "https_proxy", "all_proxy", "no_proxy"}
		for _, k := range keys {
			if v, ok := envVars[k]; ok {
				fmt.Printf("export %s=%s\n", strings.ToUpper(k), v)
			}
		}
		return nil
	},
}

var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Print unset commands to remove proxy variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"unset","data":{}}`)
			return nil
		}
		fmt.Println("unset http_proxy")
		fmt.Println("unset https_proxy")
		fmt.Println("unset all_proxy")
		fmt.Println("unset no_proxy")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(unsetCmd)
}
