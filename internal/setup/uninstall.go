package setup

import (
	"fmt"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/deps"
	"github.com/nvandessel/go4dot/internal/machine"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/nvandessel/go4dot/internal/stow"
)

// UninstallOptions configures the uninstallation behavior.
type UninstallOptions struct {
	RemoveExternal bool
	RemoveMachine  bool
	ProgressFunc   func(current, total int, msg string)
}

// Uninstall removes the dotfiles installation.
func Uninstall(cfg *config.Config, dotfilesPath string, st *state.State, opts UninstallOptions) error {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(0, 0, fmt.Sprintf("Uninstalling dotfiles from %s...", dotfilesPath))
	}

	// Get configs to unstow
	var configsToUnstow []config.ConfigItem
	if st != nil && len(st.Configs) > 0 {
		// Unstow only installed configs from state
		for _, sc := range st.Configs {
			if item := cfg.GetConfigByName(sc.Name); item != nil {
				configsToUnstow = append(configsToUnstow, *item)
			}
		}
	} else {
		// Unstow all configs from config file
		configsToUnstow = cfg.GetAllConfigs()
	}

	// Unstow configs
	if len(configsToUnstow) > 0 {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, fmt.Sprintf("Unstowing %d configs...", len(configsToUnstow)))
		}

		stowOpts := stow.StowOptions{
			ProgressFunc: opts.ProgressFunc,
		}

		result := stow.UnstowConfigs(dotfilesPath, configsToUnstow, stowOpts)

		if len(result.Failed) > 0 {
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(0, 0, fmt.Sprintf("⚠ %d configs failed to unstow", len(result.Failed)))
			}
		}
	}

	// Remove external deps if requested
	if opts.RemoveExternal && len(cfg.External) > 0 {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Removing external dependencies...")
		}

		for _, ext := range cfg.External {
			extOpts := deps.ExternalOptions{
				ProgressFunc: opts.ProgressFunc,
			}

			if err := deps.RemoveExternal(cfg, ext.ID, extOpts); err != nil {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Failed to remove %s: %v", ext.Name, err))
				}
			}
		}
	}

	// Remove machine configs if requested
	if opts.RemoveMachine && len(cfg.MachineConfig) > 0 {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Removing machine configuration files...")
		}

		for _, mc := range cfg.MachineConfig {
			renderOpts := machine.RenderOptions{
				ProgressFunc: opts.ProgressFunc,
			}

			if err := machine.RemoveMachineConfig(&mc, renderOpts); err != nil {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Failed to remove %s: %v", mc.Description, err))
				}
			}
		}
	}

	// Remove state file
	if err := state.Delete(); err != nil {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	if opts.ProgressFunc != nil {
		opts.ProgressFunc(0, 0, "✓ Removed state file")
	}

	return nil
}
