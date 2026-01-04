package main

import (
	"fmt"
	"os"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new .go4dot.yaml config",
	Long: `Scans the current directory (or provided path) for dotfiles
and generates a .go4dot.yaml configuration file interactively.

It will:
1. Scan for potential config directories (e.g. nvim, git, zsh)
2. Detect common config types
3. Prompt for project metadata
4. Generate a commented YAML file`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := config.InitConfig(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
