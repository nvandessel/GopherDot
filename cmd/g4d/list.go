package main

import (
	"fmt"
	"os"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/platform"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/nvandessel/go4dot/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [config-path]",
	Short: "List installed and available configs",
	Long: `Show the status of all dotfile configurations.

Displays:
- Installed configs (currently stowed)
- Available configs (can be installed)
- Platform-specific configs (not available on this platform)
- Archived configs (deprecated/old)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load config
		var cfg *config.Config
		var err error

		if len(args) > 0 {
			cfg, err = config.LoadFromPath(args[0])
		} else {
			cfg, _, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Load state if it exists
		st, err := state.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load state: %v\n", err)
		}

		// Detect platform
		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		showAll, _ := cmd.Flags().GetBool("all")

		ui.PrintConfigList(cfg, st, p, showAll)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolP("all", "a", false, "Show all configs including platform-specific and archived")
}
