package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/setup"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/nvandessel/go4dot/internal/ui"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [config-path]",
	Short: "Remove dotfiles installation",
	Long: `Remove all symlinks and optionally clean up external dependencies.

This command:
1. Unstows all configured dotfiles (removes symlinks)
2. Optionally removes external dependencies (--remove-external)
3. Optionally removes machine config files (--remove-machine)
4. Removes the state file

Note: This does NOT delete your dotfiles repository, only the symlinks.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load state
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

		force, _ := cmd.Flags().GetBool("force")
		removeExternal, _ := cmd.Flags().GetBool("remove-external")
		removeMachine, _ := cmd.Flags().GetBool("remove-machine")

		// Confirm unless --force
		if !force {
			fmt.Println("This will remove all dotfile symlinks from your home directory.")
			if removeExternal {
				fmt.Println("It will also remove external dependencies (plugins, themes, etc.)")
			}
			if removeMachine {
				fmt.Println("It will also remove machine-specific config files.")
			}
			fmt.Print("\nAre you sure? [y/N] ")

			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response != "y" && response != "yes" {
				fmt.Println("Aborted.")
				return
			}
			fmt.Println()
		}

		fmt.Println("Uninstalling dotfiles...")
		fmt.Printf("Directory: %s\n\n", dotfilesPath)

		opts := setup.UninstallOptions{
			RemoveExternal: removeExternal,
			RemoveMachine:  removeMachine,
			ProgressFunc: func(msg string) {
				fmt.Println("  " + msg)
			},
		}

		if err := setup.Uninstall(cfg, dotfilesPath, st, opts); err != nil {
			ui.Error("%v", err)
			os.Exit(1)
		}

		fmt.Println("\nUninstall complete!")
		fmt.Println("Your dotfiles repository is still intact at:", dotfilesPath)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	uninstallCmd.Flags().Bool("remove-external", false, "Also remove external dependencies")
	uninstallCmd.Flags().Bool("remove-machine", false, "Also remove machine-specific config files")
}
