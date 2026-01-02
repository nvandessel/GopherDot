package main

import (
	"fmt"
	"os"

	"github.com/nvandessel/gopherdot/internal/config"
	"github.com/nvandessel/gopherdot/internal/platform"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Version information (set during build)
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "gopherdot",
	Short: "GopherDot - A Go-based dotfiles manager",
	Long: `GopherDot is a CLI tool for managing dotfiles across multiple machines.
	
It provides:
  • Platform detection (OS, distro, package manager)
  • Dependency management (check and install required tools)
  • Interactive setup with beautiful TUI
  • Machine-specific configuration prompts
  • Stow-based symlink management
  • External dependency cloning (themes, plugins, etc.)
  • Health checking with doctor command
  
GopherDot works with any dotfiles repository that has a .gopherdot.yaml config file.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  "Display GopherDot version, build time, and Go version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GopherDot %s\n", Version)
		fmt.Printf("Built:      %s\n", BuildTime)
		fmt.Printf("Go version: %s\n", GoVersion)
	},
}

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect platform information",
	Long:  "Detect and display information about the current platform (OS, distro, package manager)",
	Run: func(cmd *cobra.Command, args []string) {
		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Platform Information:")
		fmt.Println("─────────────────────")
		fmt.Println(p.String())
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration files",
	Long:  "Commands for working with .gopherdot.yaml configuration files",
}

var configValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Validate a .gopherdot.yaml file",
	Long:  "Validate the syntax and structure of a .gopherdot.yaml configuration file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var configPath string
		var err error

		if len(args) > 0 {
			// Load from specified path
			configPath = args[0]
			cfg, err = config.LoadFromPath(configPath)
		} else {
			// Discover config
			cfg, configPath, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Loaded config from: %s\n", configPath)

		// Validate
		if err := cfg.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed:\n%v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Configuration is valid")
		fmt.Printf("  Schema version: %s\n", cfg.SchemaVersion)
		fmt.Printf("  Name: %s\n", cfg.Metadata.Name)
		fmt.Printf("  Configs: %d core, %d optional\n", len(cfg.Configs.Core), len(cfg.Configs.Optional))
		fmt.Printf("  Dependencies: %d total\n", len(cfg.GetAllDependencies()))
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show [path]",
	Short: "Display configuration contents",
	Long:  "Display the full contents of a .gopherdot.yaml configuration file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var configPath string
		var err error

		if len(args) > 0 {
			configPath = args[0]
			cfg, err = config.LoadFromPath(configPath)
		} else {
			cfg, configPath, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Configuration from: %s\n", configPath)
		fmt.Println("─────────────────────────────────")

		// Convert to YAML and print
		data, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
