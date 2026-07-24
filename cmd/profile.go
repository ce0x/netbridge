package main

import (
	"context"
	"fmt"

	netbridge "github.com/netbridge/netbridge"
	"github.com/spf13/cobra"
)

// resolveProfile tries Get(id) first, then GetByName(name).
func resolveProfile(ctx context.Context, mgr netbridge.ProfileManager, arg string) (*netbridge.Profile, error) {
	p, err := mgr.Get(ctx, arg)
	if err == nil {
		return p, nil
	}
	return mgr.GetByName(ctx, arg)
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Profile management commands",
}

var importCmd = &cobra.Command{
	Use:   "import <url|file>",
	Short: "Import profile from URL or file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		p, err := eng.ProfileManager().Import(cmd.Context(), args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"import","data":{"id":"%s","name":"%s","server":"%s","port":%d}}`,
				p.ID, p.Name, p.Server, p.Port)
			return nil
		}
		fmt.Printf("Imported: %s → %s:%d (id: %s)\n", p.Name, p.Server, p.Port, p.ID)
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:   "export <name>",
	Short: "Export profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		uri, err := mgr.Export(cmd.Context(), p.ID)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"export","data":{"id":"%s","uri":"%s"}}`, p.ID, uri)
			return nil
		}
		fmt.Printf("%s\n", uri)
		return nil
	},
}

var deleteProfileCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if err := mgr.Delete(cmd.Context(), p.ID); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"delete","data":{"id":"%s","name":"%s"}}`, p.ID, p.Name)
			return nil
		}
		fmt.Printf("Deleted profile: %s (id: %s)\n", p.Name, p.ID)
		return nil
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename <old> <new>",
	Short: "Rename profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if err := mgr.Rename(cmd.Context(), p.ID, args[1]); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"rename","data":{"id":"%s","old":"%s","new":"%s"}}`, p.ID, p.Name, args[1])
			return nil
		}
		fmt.Printf("Renamed %s → %s\n", p.Name, args[1])
		return nil
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone <name> <new>",
	Short: "Clone profile",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		clone, err := mgr.Clone(cmd.Context(), p.ID, args[1])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"clone","data":{"source_id":"%s","new_id":"%s","name":"%s"}}`, p.ID, clone.ID, clone.Name)
			return nil
		}
		fmt.Printf("Cloned %s → %s (id: %s)\n", p.Name, clone.Name, clone.ID)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		profiles, err := eng.ProfileManager().List(cmd.Context())
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Print(`{"success":true,"command":"list","data":{"profiles":[`)
			for i, p := range profiles {
				if i > 0 {
					fmt.Print(",")
				}
				fmt.Printf(`{"id":"%s","name":"%s","server":"%s","port":%d,"protocol":"%s"}`,
					p.ID, p.Name, p.Server, p.Port, p.Protocol)
			}
			fmt.Print(`]}}`)
			return nil
		}
		if len(profiles) == 0 {
			fmt.Println("No profiles configured.")
			return nil
		}
		for _, p := range profiles {
			fmt.Printf("  %s  %s → %s:%d\n", p.ID[:8], p.Name, p.Server, p.Port)
		}
		return nil
	},
}

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if err := mgr.SetActive(cmd.Context(), p.ID); err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"use","data":{"id":"%s","name":"%s"}}`, p.ID, p.Name)
			return nil
		}
		fmt.Printf("Active profile set to: %s (id: %s)\n", p.Name, p.ID)
		return nil
	},
}

var showCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show profile details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		eng, err := getEngine()
		if err != nil {
			return err
		}
		mgr := eng.ProfileManager()
		p, err := resolveProfile(cmd.Context(), mgr, args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		reveal, _ := cmd.Flags().GetBool("reveal")
		if jsonOutput {
			server := p.Server
			if !reveal && p.TLS.ServerName != "" {
				server = "***"
			}
			fmt.Printf(`{"success":true,"command":"show","data":{"id":"%s","name":"%s","protocol":"%s","server":"%s","port":%d,"flow":"%s"}}`,
				p.ID, p.Name, p.Protocol, server, p.Port, p.Flow)
			return nil
		}
		fmt.Printf("Profile: %s (id: %s)\n", p.Name, p.ID)
		fmt.Printf("  Protocol: %s\n", p.Protocol)
		fmt.Printf("  Server:   %s:%d\n", p.Server, p.Port)
		if p.Flow != "" {
			fmt.Printf("  Flow:     %s\n", p.Flow)
		}
		if p.Tags != nil {
			fmt.Printf("  Tags:     %v\n", p.Tags)
		}
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
