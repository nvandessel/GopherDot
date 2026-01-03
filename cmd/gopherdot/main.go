package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nvandessel/gopherdot/internal/config"
	"github.com/nvandessel/gopherdot/internal/deps"
	"github.com/nvandessel/gopherdot/internal/machine"
	"github.com/nvandessel/gopherdot/internal/platform"
	"github.com/nvandessel/gopherdot/internal/setup"
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

// Machine config commands
var machineCmd = &cobra.Command{
	Use:   "machine",
	Short: "Manage machine-specific configuration",
	Long:  "Commands for configuring machine-specific settings like git user, GPG keys, etc.",
}

var machineStatusCmd = &cobra.Command{
	Use:   "status [config-path]",
	Short: "Show status of machine configurations",
	Long:  "Display which machine-specific configurations are set up and which are missing",
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

		if len(cfg.MachineConfig) == 0 {
			fmt.Println("No machine configurations defined in config")
			return
		}

		statuses := machine.CheckMachineConfigStatus(cfg)

		fmt.Println("Machine Configuration Status")
		fmt.Println("────────────────────────────")

		var configured, missing int
		for _, s := range statuses {
			var statusIcon string
			var info string

			switch s.Status {
			case "configured":
				statusIcon = "✓"
				info = s.Destination
				configured++
			case "missing":
				statusIcon = "✗"
				info = "not configured"
				missing++
			case "error":
				statusIcon = "!"
				info = s.Error
			}

			fmt.Printf("  %s %s (%s)\n", statusIcon, s.Description, info)
		}

		fmt.Printf("\nSummary: %d configured, %d missing\n", configured, missing)

		if missing > 0 {
			fmt.Println("\nRun 'gopherdot machine configure' to set up missing configurations.")
		}
	},
}

var machineConfigureCmd = &cobra.Command{
	Use:   "configure [id] [config-path]",
	Short: "Configure machine-specific settings",
	Long: `Interactively configure machine-specific settings.

Without arguments, configures all machine settings.
With an ID argument, configures only that specific setting.`,
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

		if len(cfg.MachineConfig) == 0 {
			fmt.Println("No machine configurations defined in config")
			return
		}

		skipPrompts, _ := cmd.Flags().GetBool("defaults")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		promptOpts := machine.PromptOptions{
			SkipPrompts: skipPrompts,
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		renderOpts := machine.RenderOptions{
			Overwrite: overwrite,
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		if specificID != "" {
			// Configure single
			fmt.Printf("Configuring %s...\n\n", specificID)

			result, err := machine.CollectSingleConfig(cfg, specificID, promptOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			mc := machine.GetMachineConfigByID(cfg, specificID)
			_, err = machine.RenderAndWrite(mc, result.Values, renderOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Configure all
			fmt.Printf("Configuring %d machine settings...\n\n", len(cfg.MachineConfig))

			results, err := machine.CollectMachineConfig(cfg, promptOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			_, err = machine.RenderAll(cfg, results, renderOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println("\n✓ Configuration complete")
	},
}

var machineShowCmd = &cobra.Command{
	Use:   "show <id> [config-path]",
	Short: "Preview a machine configuration",
	Long:  "Show what a machine configuration would generate without writing it",
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

		mc := machine.GetMachineConfigByID(cfg, id)
		if mc == nil {
			fmt.Fprintf(os.Stderr, "Error: machine config '%s' not found\n", id)
			os.Exit(1)
		}

		// Collect values (use defaults)
		result, err := machine.CollectSingleConfig(cfg, id, machine.PromptOptions{SkipPrompts: true})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting defaults: %v\n", err)
			os.Exit(1)
		}

		content, err := machine.PreviewRender(mc, result.Values)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering preview: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Preview of %s (destination: %s):\n", mc.Description, mc.Destination)
		fmt.Println("────────────────────────────────────")
		fmt.Println(content)
	},
}

var machineRemoveCmd = &cobra.Command{
	Use:   "remove <id> [config-path]",
	Short: "Remove a machine configuration file",
	Long:  "Remove a generated machine configuration file",
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

		mc := machine.GetMachineConfigByID(cfg, id)
		if mc == nil {
			fmt.Fprintf(os.Stderr, "Error: machine config '%s' not found\n", id)
			os.Exit(1)
		}

		opts := machine.RenderOptions{
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		err = machine.RemoveMachineConfig(mc, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var machineInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show system information for machine config",
	Long:  "Display detected system information useful for machine configuration",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := machine.GetSystemInfo()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting system info: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("System Information")
		fmt.Println("──────────────────")
		fmt.Printf("Username:   %s\n", info.Username)
		fmt.Printf("Hostname:   %s\n", info.Hostname)
		fmt.Printf("Home:       %s\n", info.HomeDir)
		fmt.Println()

		fmt.Println("Git Configuration")
		fmt.Println("─────────────────")
		if info.GitUserName != "" {
			fmt.Printf("user.name:  %s\n", info.GitUserName)
		} else {
			fmt.Println("user.name:  (not configured)")
		}
		if info.GitEmail != "" {
			fmt.Printf("user.email: %s\n", info.GitEmail)
		} else {
			fmt.Println("user.email: (not configured)")
		}
		fmt.Println()

		fmt.Println("Security Keys")
		fmt.Println("─────────────")
		if info.HasGPG {
			fmt.Println("GPG:        ✓ Keys available")
			keys, _ := machine.DetectGPGKeys()
			for _, key := range keys {
				fmt.Printf("            - %s (%s)\n", key.Email, key.KeyID)
			}
		} else {
			fmt.Println("GPG:        ✗ No keys found")
		}

		if info.HasSSH {
			fmt.Println("SSH:        ✓ Keys loaded")
			keys, _ := machine.DetectSSHKeys()
			for _, key := range keys {
				fmt.Printf("            - %s (%s)\n", key.Path, key.Type)
			}
		} else {
			fmt.Println("SSH:        ✗ No keys loaded in agent")
		}
	},
}

// Install command - main entry point for setting up dotfiles
var installCmd = &cobra.Command{
	Use:   "install [config-path]",
	Short: "Install and configure dotfiles",
	Long: `Run the full dotfiles installation process.

This command orchestrates:
1. Dependency checking and installation
2. Stowing dotfile configurations
3. Cloning external dependencies (plugins, themes)
4. Configuring machine-specific settings

Use flags to customize the installation:
  --auto       Non-interactive mode, use defaults
  --minimal    Only install core configs
  --skip-deps  Skip dependency installation
  --skip-external  Skip external dependency cloning
  --skip-machine   Skip machine-specific configuration
  --skip-stow      Skip stowing configs`,
	Args: cobra.MaximumNArgs(1),
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

		// Get flags
		auto, _ := cmd.Flags().GetBool("auto")
		minimal, _ := cmd.Flags().GetBool("minimal")
		skipDeps, _ := cmd.Flags().GetBool("skip-deps")
		skipExternal, _ := cmd.Flags().GetBool("skip-external")
		skipMachine, _ := cmd.Flags().GetBool("skip-machine")
		skipStow, _ := cmd.Flags().GetBool("skip-stow")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		opts := setup.InstallOptions{
			Auto:         auto,
			Minimal:      minimal,
			SkipDeps:     skipDeps,
			SkipExternal: skipExternal,
			SkipMachine:  skipMachine,
			SkipStow:     skipStow,
			Overwrite:    overwrite,
			ProgressFunc: func(msg string) {
				fmt.Println(msg)
			},
		}

		// Print header
		fmt.Println("╔════════════════════════════════════════╗")
		fmt.Println("║        GopherDot Installation          ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\nDotfiles: %s\n", dotfilesPath)
		if cfg.Metadata.Name != "" {
			fmt.Printf("Config:   %s\n", cfg.Metadata.Name)
		}
		fmt.Println()

		result, err := setup.Install(cfg, dotfilesPath, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			os.Exit(1)
		}

		// Print summary
		fmt.Println("\n════════════════════════════════════════")
		if result.HasErrors() {
			fmt.Println("Installation completed with errors")
			fmt.Println()
			fmt.Print(result.Summary())

			// Show specific errors
			for _, e := range result.DepsFailed {
				fmt.Printf("  ✗ Dependency %s: %v\n", e.Item.Name, e.Error)
			}
			for _, e := range result.ConfigsFailed {
				fmt.Printf("  ✗ Config %s: %v\n", e.ConfigName, e.Error)
			}
			for _, e := range result.ExternalFailed {
				fmt.Printf("  ✗ External %s: %v\n", e.Dep.Name, e.Error)
			}
			for _, e := range result.Errors {
				fmt.Printf("  ✗ %v\n", e)
			}
			os.Exit(1)
		} else {
			fmt.Println("✓ Installation complete!")
			fmt.Println()
			fmt.Print(result.Summary())

			// Show post-install message if present
			if cfg.PostInstall != "" {
				fmt.Println("\n── Next Steps ──")
				fmt.Println(cfg.PostInstall)
			}
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
	rootCmd.AddCommand(machineCmd)
	rootCmd.AddCommand(installCmd)

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

	machineCmd.AddCommand(machineStatusCmd)
	machineCmd.AddCommand(machineConfigureCmd)
	machineCmd.AddCommand(machineShowCmd)
	machineCmd.AddCommand(machineRemoveCmd)
	machineCmd.AddCommand(machineInfoCmd)

	// Flags for machine configure
	machineConfigureCmd.Flags().Bool("defaults", false, "Use default values without prompting")
	machineConfigureCmd.Flags().Bool("overwrite", false, "Overwrite existing configuration files")

	// Flags for install
	installCmd.Flags().Bool("auto", false, "Non-interactive mode, use defaults")
	installCmd.Flags().Bool("minimal", false, "Only install core configs, skip optional")
	installCmd.Flags().Bool("skip-deps", false, "Skip dependency installation")
	installCmd.Flags().Bool("skip-external", false, "Skip external dependency cloning")
	installCmd.Flags().Bool("skip-machine", false, "Skip machine-specific configuration")
	installCmd.Flags().Bool("skip-stow", false, "Skip stowing configs")
	installCmd.Flags().Bool("overwrite", false, "Overwrite existing files")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
