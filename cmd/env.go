package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print export commands for proxy environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"env","data":{"http_proxy":"http://127.0.0.1:8080","https_proxy":"http://127.0.0.1:8080","all_proxy":"socks5://127.0.0.1:10808"}}`)
			return nil
		}
		fmt.Println("export http_proxy=http://127.0.0.1:8080")
		fmt.Println("export https_proxy=http://127.0.0.1:8080")
		fmt.Println("export all_proxy=socks5://127.0.0.1:10808")
		fmt.Println("export no_proxy=localhost,127.0.0.1,::1")
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
