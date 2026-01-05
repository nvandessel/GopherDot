package ui

import (
	"fmt"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/platform"
	"github.com/nvandessel/go4dot/internal/state"
)

// PrintConfigList prints the status of all configs using internal/ui styles.
func PrintConfigList(cfg *config.Config, st *state.State, p *platform.Platform, showAll bool) {
	// Build installed set from state
	installed := make(map[string]bool)
	if st != nil {
		for _, c := range st.Configs {
			installed[c.Name] = true
		}
	}

	// Core configs
	Section("Core Configs")
	for _, c := range cfg.Configs.Core {
		printConfigStatus(c, installed, p, showAll)
	}

	// Optional configs
	if len(cfg.Configs.Optional) > 0 {
		Section("Optional Configs")
		for _, c := range cfg.Configs.Optional {
			printConfigStatus(c, installed, p, showAll)
		}
	}

	// External deps
	if len(cfg.External) > 0 {
		Section("External Dependencies")
		for _, e := range cfg.External {
			status := "x"
			info := "not installed"

			if st != nil {
				if ext, ok := st.ExternalDeps[e.ID]; ok && ext.Installed {
					status = "+"
					info = ext.Path
				}
			}

			// Check if skipped due to platform
			if !platform.CheckCondition(e.Condition, p) {
				if showAll {
					fmt.Printf("  ⊘ %s (skipped - platform mismatch)\n", e.Name)
				}
				continue
			}

			if status == "+" {
				Success("%s (%s)", e.Name, info)
			} else {
				fmt.Printf("  • %s (%s)\n", e.Name, info)
			}
		}
	}

	// Machine configs
	if len(cfg.MachineConfig) > 0 {
		Section("Machine Configurations")
		for _, mc := range cfg.MachineConfig {
			status := "x"
			info := "not configured"

			if st != nil {
				if m, ok := st.MachineConfig[mc.ID]; ok {
					status = "+"
					info = m.ConfigPath
				}
			}

			if status == "+" {
				Success("%s (%s)", mc.Description, info)
			} else {
				fmt.Printf("  • %s (%s)\n", mc.Description, info)
			}
		}
	}

	// Archived configs
	if len(cfg.Archived) > 0 && showAll {
		Section("Archived Configs (deprecated)")
		for _, c := range cfg.Archived {
			fmt.Printf("  - %s\n", c.Name)
			if c.Description != "" {
				fmt.Printf("    %s\n", c.Description)
			}
		}
	}

	// Summary
	Section("Summary")
	if st != nil {
		fmt.Printf("Installed: %d configs\n", len(st.Configs))
		if st.DotfilesPath != "" {
			fmt.Printf("Dotfiles:  %s\n", st.DotfilesPath)
		}
	} else {
		Warning("No installation state found. Run 'g4d install' to set up.")
	}
}

func printConfigStatus(c config.ConfigItem, installed map[string]bool, p *platform.Platform, showAll bool) {
	// Check platform compatibility
	if len(c.Platforms) > 0 && !isPlatformMatch(c.Platforms, p) {
		if showAll {
			fmt.Printf("  ⊘ %s (not available on %s)\n", c.Name, p.OS)
		}
		return
	}

	if installed[c.Name] {
		Success("%s - %s (installed)", c.Name, c.Description)
	} else {
		fmt.Printf("  • %s - %s (not installed)\n", c.Name, c.Description)
	}
}

func isPlatformMatch(platforms []string, p *platform.Platform) bool {
	for _, plat := range platforms {
		if plat == p.OS || plat == "all" {
			return true
		}
		// Also check distro
		if plat == p.Distro {
			return true
		}
	}
	return false
}
