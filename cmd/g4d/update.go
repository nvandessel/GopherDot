package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/setup"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/nvandessel/go4dot/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [config-path]",
	Short: "Update dotfiles from git",
	Long: `Pull latest changes from git and update dotfiles.

This command:
1. Runs git pull in the dotfiles directory
2. Shows what files changed
3. Restows all configs to apply changes
4. Updates external dependencies (if --external flag is set)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load state to get dotfiles path
		st, err := state.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}

		var dotfilesPath string
		var cfg *config.Config

		if len(args) > 0 {
			cfg, err = config.LoadFromPath(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			dotfilesPath = filepath.Dir(args[0])
		} else if st != nil && st.DotfilesPath != "" {
			dotfilesPath = st.DotfilesPath
			cfg, err = config.LoadFromPath(dotfilesPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
		} else {
			cfg, dotfilesPath, err = config.LoadFromDiscovery()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			dotfilesPath = filepath.Dir(dotfilesPath)
		}

		updateExternal, _ := cmd.Flags().GetBool("external")
		skipRestow, _ := cmd.Flags().GetBool("skip-restow")

		fmt.Println("Updating dotfiles...")
		fmt.Printf("Directory: %s\n\n", dotfilesPath)

		opts := setup.UpdateOptions{
			UpdateExternal: updateExternal,
			SkipRestow:     skipRestow,
			ProgressFunc: func(msg string) {
				fmt.Println("  " + msg)
			},
		}

		if err := setup.Update(cfg, dotfilesPath, st, opts); err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}

		fmt.Println("\nUpdate complete!")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("external", false, "Also update external dependencies")
	updateCmd.Flags().Bool("skip-restow", false, "Skip restowing configs after pull")
}
