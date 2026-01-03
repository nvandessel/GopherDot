package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nvandessel/gopherdot/internal/config"
	"github.com/nvandessel/gopherdot/internal/deps"
	"github.com/nvandessel/gopherdot/internal/platform"
	"github.com/nvandessel/gopherdot/internal/stow"
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

var depsCmd = &cobra.Command{
	Use:   "deps",
	Short: "Manage dependencies",
	Long:  "Commands for checking and installing system dependencies",
}

var depsCheckCmd = &cobra.Command{
	Use:   "check [config-path]",
	Short: "Check dependency status",
	Long:  "Check which dependencies are installed and which are missing",
	Args:  cobra.MaximumNArgs(1),
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

		// Detect platform
		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		// Check dependencies
		result, err := deps.Check(cfg, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
			os.Exit(1)
		}

		// Display results
		fmt.Println("Dependency Status")
		fmt.Println("─────────────────")
		fmt.Printf("Package Manager: %s\n", p.PackageManager)
		fmt.Printf("Summary: %s\n\n", result.Summary())

		// Show critical deps
		if len(result.Critical) > 0 {
			fmt.Println("Critical Dependencies:")
			for _, dep := range result.Critical {
				printDepStatus(dep)
			}
			fmt.Println()
		}

		// Show core deps
		if len(result.Core) > 0 {
			fmt.Println("Core Dependencies:")
			for _, dep := range result.Core {
				printDepStatus(dep)
			}
			fmt.Println()
		}

		// Show optional deps
		if len(result.Optional) > 0 {
			fmt.Println("Optional Dependencies:")
			for _, dep := range result.Optional {
				printDepStatus(dep)
			}
		}

		// Exit with error if critical deps are missing
		if len(result.GetMissingCritical()) > 0 {
			fmt.Fprintf(os.Stderr, "\nError: Missing critical dependencies. Run 'gopherdot deps install' to install them.\n")
			os.Exit(1)
		}
	},
}

var depsInstallCmd = &cobra.Command{
	Use:   "install [config-path]",
	Short: "Install missing dependencies",
	Long:  "Install system packages for missing dependencies",
	Args:  cobra.MaximumNArgs(1),
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

		// Detect platform
		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		// Check current status
		checkResult, err := deps.Check(cfg, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
			os.Exit(1)
		}

		missing := checkResult.GetMissing()
		if len(missing) == 0 {
			fmt.Println("✓ All dependencies are already installed!")
			return
		}

		fmt.Printf("Installing %d missing dependencies...\n\n", len(missing))

		// Install with progress
		opts := deps.InstallOptions{
			OnlyMissing: true,
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		result, err := deps.Install(cfg, p, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during installation: %v\n", err)
			os.Exit(1)
		}

		// Show results
		fmt.Println()
		fmt.Printf("✓ Installed: %d packages\n", len(result.Installed))
		if len(result.Failed) > 0 {
			fmt.Printf("✗ Failed: %d packages\n", len(result.Failed))
			for _, fail := range result.Failed {
				fmt.Printf("  - %s: %v\n", fail.Item.Name, fail.Error)
			}
			os.Exit(1)
		}
	},
}

func printDepStatus(dep deps.DependencyCheck) {
	status := "✗"
	info := "missing"

	if dep.Status == deps.StatusInstalled {
		status = "✓"
		info = dep.InstalledPath
	}

	fmt.Printf("  %s %s (%s)\n", status, dep.Item.Name, info)
}

var stowCmd = &cobra.Command{
	Use:   "stow",
	Short: "Manage dotfile symlinks",
	Long:  "Commands for stowing, unstowing, and managing dotfile symlinks",
}

var stowAddCmd = &cobra.Command{
	Use:   "add <config-name> [config-path]",
	Short: "Stow a specific config",
	Long:  "Create symlinks for a specific dotfile configuration",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]

		// Load config
		var cfg *config.Config
		var configPath string
		var err error

		if len(args) > 1 {
			cfg, err = config.LoadFromPath(args[1])
			configPath = args[1]
		} else {
			cfg, configPath, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Find the config item
		cfgItem := cfg.GetConfigByName(configName)
		if cfgItem == nil {
			fmt.Fprintf(os.Stderr, "Error: config '%s' not found\n", configName)
			os.Exit(1)
		}

		// Get dotfiles directory
		dotfilesPath := filepath.Dir(configPath)

		// Stow it
		opts := stow.StowOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		err = stow.Stow(dotfilesPath, cfgItem.Path, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var stowRemoveCmd = &cobra.Command{
	Use:   "remove <config-name> [config-path]",
	Short: "Unstow a specific config",
	Long:  "Remove symlinks for a specific dotfile configuration",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		configName := args[0]

		var cfg *config.Config
		var configPath string
		var err error

		if len(args) > 1 {
			cfg, err = config.LoadFromPath(args[1])
			configPath = args[1]
		} else {
			cfg, configPath, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		cfgItem := cfg.GetConfigByName(configName)
		if cfgItem == nil {
			fmt.Fprintf(os.Stderr, "Error: config '%s' not found\n", configName)
			os.Exit(1)
		}

		dotfilesPath := filepath.Dir(configPath)

		opts := stow.StowOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		err = stow.Unstow(dotfilesPath, cfgItem.Path, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var stowRefreshCmd = &cobra.Command{
	Use:   "refresh [config-path]",
	Short: "Refresh all stowed configs",
	Long:  "Restow all configs to update symlinks",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var configPath string
		var err error

		if len(args) > 0 {
			cfg, err = config.LoadFromPath(args[0])
			configPath = args[0]
		} else {
			cfg, configPath, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		dotfilesPath := filepath.Dir(configPath)

		// Restow all configs
		opts := stow.StowOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		allConfigs := cfg.GetAllConfigs()
		fmt.Printf("Refreshing %d configs...\n\n", len(allConfigs))

		result := stow.RestowConfigs(dotfilesPath, allConfigs, opts)

		// Show results
		fmt.Println()
		if len(result.Success) > 0 {
			fmt.Printf("✓ Refreshed: %d configs\n", len(result.Success))
		}
		if len(result.Skipped) > 0 {
			fmt.Printf("⊘ Skipped: %d configs\n", len(result.Skipped))
		}
		if len(result.Failed) > 0 {
			fmt.Printf("✗ Failed: %d configs\n", len(result.Failed))
			for _, fail := range result.Failed {
				fmt.Printf("  - %s: %v\n", fail.ConfigName, fail.Error)
			}
			os.Exit(1)
		}
	},
}

// External dependency commands
var externalCmd = &cobra.Command{
	Use:   "external",
	Short: "Manage external dependencies",
	Long:  "Commands for cloning, updating, and managing external dependencies (plugins, themes, etc.)",
}

var externalStatusCmd = &cobra.Command{
	Use:   "status [config-path]",
	Short: "Show status of external dependencies",
	Long:  "Display the installation status of all external dependencies",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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

		if len(cfg.External) == 0 {
			fmt.Println("No external dependencies defined in config")
			return
		}

		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		statuses := deps.CheckExternalStatus(cfg, p)

		fmt.Println("External Dependencies Status")
		fmt.Println("────────────────────────────")

		var installed, missing, skipped int
		for _, s := range statuses {
			var statusIcon string
			var info string

			switch s.Status {
			case "installed":
				statusIcon = "✓"
				info = s.Path
				installed++
			case "missing":
				statusIcon = "✗"
				info = "not installed"
				missing++
			case "skipped":
				statusIcon = "⊘"
				info = s.Reason
				skipped++
			case "error":
				statusIcon = "!"
				info = s.Reason
			}

			fmt.Printf("  %s %s (%s)\n", statusIcon, s.Dep.Name, info)
		}

		fmt.Printf("\nSummary: %d installed, %d missing, %d skipped\n", installed, missing, skipped)

		if missing > 0 {
			fmt.Println("\nRun 'gopherdot external clone' to install missing dependencies.")
		}
	},
}

var externalCloneCmd = &cobra.Command{
	Use:   "clone [id] [config-path]",
	Short: "Clone external dependencies",
	Long: `Clone external dependencies from their repositories.

Without arguments, clones all missing external dependencies.
With an ID argument, clones only that specific dependency.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var err error
		var specificID string

		// Parse arguments
		configPathArg := ""
		if len(args) >= 1 {
			// Could be ID or config path
			// If it looks like a path, treat it as such
			if _, statErr := os.Stat(args[0]); statErr == nil || filepath.Ext(args[0]) == ".yaml" || filepath.Ext(args[0]) == ".yml" {
				configPathArg = args[0]
			} else {
				specificID = args[0]
				if len(args) >= 2 {
					configPathArg = args[1]
				}
			}
		}

		if configPathArg != "" {
			cfg, err = config.LoadFromPath(configPathArg)
		} else {
			cfg, _, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.External) == 0 {
			fmt.Println("No external dependencies defined in config")
			return
		}

		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		opts := deps.ExternalOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		if specificID != "" {
			// Clone single
			fmt.Printf("Cloning %s...\n\n", specificID)
			err = deps.CloneSingle(cfg, p, specificID, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n✓ Done")
		} else {
			// Clone all
			fmt.Printf("Cloning %d external dependencies...\n\n", len(cfg.External))
			result, err := deps.CloneExternal(cfg, p, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Show results
			fmt.Println()
			if len(result.Cloned) > 0 {
				fmt.Printf("✓ Cloned: %d\n", len(result.Cloned))
			}
			if len(result.Updated) > 0 {
				fmt.Printf("↻ Updated: %d\n", len(result.Updated))
			}
			if len(result.Skipped) > 0 {
				fmt.Printf("⊘ Skipped: %d\n", len(result.Skipped))
			}
			if len(result.Failed) > 0 {
				fmt.Printf("✗ Failed: %d\n", len(result.Failed))
				for _, fail := range result.Failed {
					fmt.Printf("  - %s: %v\n", fail.Dep.Name, fail.Error)
				}
				os.Exit(1)
			}
		}
	},
}

var externalUpdateCmd = &cobra.Command{
	Use:   "update [id] [config-path]",
	Short: "Update external dependencies",
	Long: `Pull updates for installed external dependencies.

Without arguments, updates all installed external dependencies.
With an ID argument, updates only that specific dependency.`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var cfg *config.Config
		var err error
		var specificID string

		configPathArg := ""
		if len(args) >= 1 {
			if _, statErr := os.Stat(args[0]); statErr == nil || filepath.Ext(args[0]) == ".yaml" || filepath.Ext(args[0]) == ".yml" {
				configPathArg = args[0]
			} else {
				specificID = args[0]
				if len(args) >= 2 {
					configPathArg = args[1]
				}
			}
		}

		if configPathArg != "" {
			cfg, err = config.LoadFromPath(configPathArg)
		} else {
			cfg, _, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.External) == 0 {
			fmt.Println("No external dependencies defined in config")
			return
		}

		p, err := platform.Detect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error detecting platform: %v\n", err)
			os.Exit(1)
		}

		opts := deps.ExternalOptions{
			Update: true,
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		if specificID != "" {
			// Update single
			fmt.Printf("Updating %s...\n\n", specificID)
			err = deps.CloneSingle(cfg, p, specificID, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("\n✓ Done")
		} else {
			// Update all
			fmt.Printf("Updating %d external dependencies...\n\n", len(cfg.External))
			result, err := deps.CloneExternal(cfg, p, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Show results
			fmt.Println()
			if len(result.Updated) > 0 {
				fmt.Printf("↻ Updated: %d\n", len(result.Updated))
			}
			if len(result.Cloned) > 0 {
				fmt.Printf("✓ Cloned (new): %d\n", len(result.Cloned))
			}
			if len(result.Skipped) > 0 {
				fmt.Printf("⊘ Skipped: %d\n", len(result.Skipped))
			}
			if len(result.Failed) > 0 {
				fmt.Printf("✗ Failed: %d\n", len(result.Failed))
				for _, fail := range result.Failed {
					fmt.Printf("  - %s: %v\n", fail.Dep.Name, fail.Error)
				}
				os.Exit(1)
			}
		}
	},
}

var externalRemoveCmd = &cobra.Command{
	Use:   "remove <id> [config-path]",
	Short: "Remove an external dependency",
	Long:  "Remove an installed external dependency by its ID",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]

		var cfg *config.Config
		var err error

		if len(args) > 1 {
			cfg, err = config.LoadFromPath(args[1])
		} else {
			cfg, _, err = config.LoadFromDiscovery()
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		opts := deps.ExternalOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		err = deps.RemoveExternal(cfg, id, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(detectCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(stowCmd)
	rootCmd.AddCommand(externalCmd)

	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)

	depsCmd.AddCommand(depsCheckCmd)
	depsCmd.AddCommand(depsInstallCmd)

	stowCmd.AddCommand(stowAddCmd)
	stowCmd.AddCommand(stowRemoveCmd)
	stowCmd.AddCommand(stowRefreshCmd)

	externalCmd.AddCommand(externalStatusCmd)
	externalCmd.AddCommand(externalCloneCmd)
	externalCmd.AddCommand(externalUpdateCmd)
	externalCmd.AddCommand(externalRemoveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
