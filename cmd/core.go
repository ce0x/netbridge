package main

import (
	"fmt"

	"github.com/netbridge/netbridge/internal/corebin"
	"github.com/spf13/cobra"
)

var coreCmd = &cobra.Command{
	Use:   "core",
	Short: "Manage core binaries (xray, sing-box, wireguard-tools, openvpn)",
}

var coreInstallCmd = &cobra.Command{
	Use:   "install [name|--all]",
	Short: "Install core binary",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		mgr := corebin.NewManager()

		if all {
			for _, c := range mgr.ListCores() {
				installed, _, _ := mgr.Detect(c.Name)
				if installed {
					fmt.Printf("  %s already installed\n", c.Name)
					continue
				}
				if c.Source == "github" {
					gh := corebin.NewGitHubInstaller(mgr.InstallDir())
					if err := gh.Install(cmd.Context(), c.Name); err != nil {
						fmt.Printf("  %s: install failed: %v\n", c.Name, err)
					}
				} else {
					pm := corebin.NewPackageManagerInstaller()
					if err := pm.Install(cmd.Context(), c.Name); err != nil {
						fmt.Printf("  %s: install failed: %v\n", c.Name, err)
					}
				}
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify core name or use --all")
		}

		name := args[0]
		def, err := mgr.GetCoreDef(name)
		if err != nil {
			return err
		}

		if def.Source == "github" {
			gh := corebin.NewGitHubInstaller(mgr.InstallDir())
			return gh.Install(cmd.Context(), name)
		}
		pm := corebin.NewPackageManagerInstaller()
		return pm.Install(cmd.Context(), name)
	},
}

var coreStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all core binaries",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := corebin.NewManager()

		if jsonOutput {
			fmt.Print(`{"success":true,"command":"core_status","data":{"cores":[`)
			cores := mgr.ListCores()
			for i, c := range cores {
				if i > 0 {
					fmt.Print(",")
				}
				installed, version, path := mgr.Detect(c.Name)
				fmt.Printf(`{"name":"%s","installed":%v,"version":"%s","path":"%s","source":"%s"}`,
					c.Name, installed, version, path, c.Source)
			}
			fmt.Print(`]}}`)
			return nil
		}

		cores := mgr.ListCores()
		fmt.Printf("%-20s %-10s %-30s %s\n", "Name", "Status", "Version", "Source")
		fmt.Println("─────────────────────────────────────────────────────────────")
		for _, c := range cores {
			installed, version, _ := mgr.Detect(c.Name)
			status := "✗ missing"
			if installed {
				status = "✓ installed"
			}
			fmt.Printf("%-20s %-10s %-30s %s\n", c.Name, status, version, c.Source)
		}
		return nil
	},
}

var coreUpdateCmd = &cobra.Command{
	Use:   "update [name|--all]",
	Short: "Update core binary to latest version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		mgr := corebin.NewManager()

		if all {
			for _, c := range mgr.ListCores() {
				if c.Source == "github" {
					gh := corebin.NewGitHubInstaller(mgr.InstallDir())
					if err := gh.Update(cmd.Context(), c.Name); err != nil {
						fmt.Printf("  %s: update failed: %v\n", c.Name, err)
					}
				} else {
					pm := corebin.NewPackageManagerInstaller()
					if err := pm.Update(cmd.Context(), c.Name); err != nil {
						fmt.Printf("  %s: update failed: %v\n", c.Name, err)
					}
				}
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("specify core name or use --all")
		}

		name := args[0]
		def, err := mgr.GetCoreDef(name)
		if err != nil {
			return err
		}

		if def.Source == "github" {
			gh := corebin.NewGitHubInstaller(mgr.InstallDir())
			return gh.Update(cmd.Context(), name)
		}
		pm := corebin.NewPackageManagerInstaller()
		return pm.Update(cmd.Context(), name)
	},
}

var coreRepairCmd = &cobra.Command{
	Use:   "repair [name]",
	Short: "Repair corrupted core binary",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr := corebin.NewManager()
		def, err := mgr.GetCoreDef(name)
		if err != nil {
			return err
		}

		if def.Source == "github" {
			gh := corebin.NewGitHubInstaller(mgr.InstallDir())
			return gh.Repair(cmd.Context(), name)
		}
		pm := corebin.NewPackageManagerInstaller()
		return pm.Repair(cmd.Context(), name)
	},
}

func init() {
	coreInstallCmd.Flags().Bool("all", false, "install all cores")
	coreUpdateCmd.Flags().Bool("all", false, "update all cores")

	coreCmd.AddCommand(coreInstallCmd, coreStatusCmd, coreUpdateCmd, coreRepairCmd)
	rootCmd.AddCommand(coreCmd)
}
