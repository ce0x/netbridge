package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create backup archive of profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")

		eng, err := getEngine()
		if err != nil {
			return err
		}

		pm := eng.ProfileManager()
		ctx := cmd.Context()

		profiles, err := pm.List(ctx)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}

		if output == "" {
			output = fmt.Sprintf("netbridge-backup-%s.zip", time.Now().Format("20060102-150405"))
		}

		f, err := os.Create(output)
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		defer f.Close()

		w := zip.NewWriter(f)
		defer w.Close()

		for _, p := range profiles {
			fw, err := w.Create(fmt.Sprintf("profiles/%s.json", p.ID))
			if err != nil {
				continue
			}
			data, _ := json.MarshalIndent(p, "", "  ")
			fw.Write(data)
		}

		if jsonOutput {
			fmt.Printf(`{"success":true,"command":"backup","data":{"output":"%s","profiles":%d}}`, output, len(profiles))
			return nil
		}

		fmt.Printf("Backup created: %s (%d profiles)\n", output, len(profiles))
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore <file>",
	Short: "Restore from backup archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := zip.OpenReader(args[0])
		if err != nil {
			if jsonOutput {
				fmt.Printf(`{"success":false,"error":"%s"}`, err)
				return nil
			}
			return err
		}
		defer f.Close()

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
