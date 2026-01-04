package setup

import (
	"fmt"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/deps"
	"github.com/nvandessel/go4dot/internal/machine"
	"github.com/nvandessel/go4dot/internal/platform"
	"github.com/nvandessel/go4dot/internal/stow"
)

// InstallOptions configures the installation behavior
type InstallOptions struct {
	Auto         bool             // Non-interactive, use defaults
	Minimal      bool             // Only core configs, skip optional
	SkipDeps     bool             // Skip dependency installation
	SkipExternal bool             // Skip external dependency cloning
	SkipMachine  bool             // Skip machine-specific configuration
	SkipStow     bool             // Skip stowing configs
	Overwrite    bool             // Overwrite existing files
	ProgressFunc func(msg string) // Called for progress updates
}

// InstallResult tracks the result of the installation
type InstallResult struct {
	Platform       *platform.Platform
	DepsInstalled  []config.DependencyItem
	DepsFailed     []deps.InstallError
	ConfigsStowed  []string
	ConfigsFailed  []stow.StowError
	ExternalCloned []config.ExternalDep
	ExternalFailed []deps.ExternalError
	MachineConfigs []machine.RenderResult
	Errors         []error
}

// HasErrors returns true if any errors occurred during installation
func (r *InstallResult) HasErrors() bool {
	return len(r.DepsFailed) > 0 || len(r.ConfigsFailed) > 0 ||
		len(r.ExternalFailed) > 0 || len(r.Errors) > 0
}

// Install runs the full installation flow
func Install(cfg *config.Config, dotfilesPath string, opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{}

	// Step 1: Detect platform
	progress(opts, "Detecting platform...")
	p, err := platform.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect platform: %w", err)
	}
	result.Platform = p
	progress(opts, fmt.Sprintf("✓ Platform: %s (%s)", p.OS, p.PackageManager))

	// Step 2: Check and install dependencies
	if !opts.SkipDeps {
		if err := installDependencies(cfg, p, opts, result); err != nil {
			result.Errors = append(result.Errors, err)
			// Don't return - continue with other steps
		}
	} else {
		progress(opts, "⊘ Skipping dependency installation")
	}

	// Step 3: Stow configs
	if !opts.SkipStow {
		if err := stowConfigs(cfg, dotfilesPath, opts, result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	} else {
		progress(opts, "⊘ Skipping config stowing")
	}

	// Step 4: Clone external dependencies
	if !opts.SkipExternal {
		if err := cloneExternal(cfg, dotfilesPath, p, opts, result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	} else {
		progress(opts, "⊘ Skipping external dependencies")
	}

	// Step 5: Configure machine-specific settings
	if !opts.SkipMachine {
		if err := configureMachine(cfg, opts, result); err != nil {
			result.Errors = append(result.Errors, err)
		}
	} else {
		progress(opts, "⊘ Skipping machine configuration")
	}

	return result, nil
}

// installDependencies checks and installs missing dependencies
func installDependencies(cfg *config.Config, p *platform.Platform, opts InstallOptions, result *InstallResult) error {
	progress(opts, "\n── Dependencies ──")

	// Check current status
	checkResult, err := deps.Check(cfg, p)
	if err != nil {
		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	missing := checkResult.GetMissing()
	if len(missing) == 0 {
		progress(opts, "✓ All dependencies are installed")
		return nil
	}

	progress(opts, fmt.Sprintf("Installing %d missing dependencies...", len(missing)))

	installOpts := deps.InstallOptions{
		OnlyMissing: true,
		ProgressFunc: func(msg string) {
			progress(opts, "  "+msg)
		},
	}

	installResult, err := deps.Install(cfg, p, installOpts)
	if err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	result.DepsInstalled = installResult.Installed
	result.DepsFailed = installResult.Failed

	if len(installResult.Failed) > 0 {
		progress(opts, fmt.Sprintf("⚠ %d dependencies failed to install", len(installResult.Failed)))
	} else {
		progress(opts, fmt.Sprintf("✓ Installed %d dependencies", len(installResult.Installed)))
	}

	return nil
}

// stowConfigs stows all or selected configs
func stowConfigs(cfg *config.Config, dotfilesPath string, opts InstallOptions, result *InstallResult) error {
	progress(opts, "\n── Configs ──")

	// Get configs to stow
	var configs []config.ConfigItem
	if opts.Minimal {
		configs = cfg.Configs.Core
	} else {
		configs = cfg.GetAllConfigs()
	}

	if len(configs) == 0 {
		progress(opts, "No configs to stow")
		return nil
	}

	progress(opts, fmt.Sprintf("Stowing %d configs...", len(configs)))

	stowOpts := stow.StowOptions{
		ProgressFunc: func(msg string) {
			progress(opts, "  "+msg)
		},
	}

	stowResult := stow.StowConfigs(dotfilesPath, configs, stowOpts)

	result.ConfigsStowed = stowResult.Success
	result.ConfigsFailed = stowResult.Failed

	if len(stowResult.Failed) > 0 {
		progress(opts, fmt.Sprintf("⚠ %d configs failed to stow", len(stowResult.Failed)))
	}
	if len(stowResult.Success) > 0 {
		progress(opts, fmt.Sprintf("✓ Stowed %d configs", len(stowResult.Success)))
	}
	if len(stowResult.Skipped) > 0 {
		progress(opts, fmt.Sprintf("⊘ Skipped %d configs (not found)", len(stowResult.Skipped)))
	}

	return nil
}

// cloneExternal clones external dependencies
func cloneExternal(cfg *config.Config, dotfilesPath string, p *platform.Platform, opts InstallOptions, result *InstallResult) error {
	if len(cfg.External) == 0 {
		return nil
	}

	progress(opts, "\n── External Dependencies ──")
	progress(opts, fmt.Sprintf("Cloning %d external dependencies...", len(cfg.External)))

	extOpts := deps.ExternalOptions{
		RepoRoot: dotfilesPath,
		ProgressFunc: func(msg string) {
			progress(opts, "  "+msg)
		},
	}

	extResult, err := deps.CloneExternal(cfg, p, extOpts)
	if err != nil {
		return fmt.Errorf("failed to clone external dependencies: %w", err)
	}

	result.ExternalCloned = extResult.Cloned
	result.ExternalFailed = extResult.Failed

	if len(extResult.Failed) > 0 {
		progress(opts, fmt.Sprintf("⚠ %d external deps failed", len(extResult.Failed)))
	}
	if len(extResult.Cloned) > 0 {
		progress(opts, fmt.Sprintf("✓ Cloned %d external deps", len(extResult.Cloned)))
	}
	if len(extResult.Skipped) > 0 {
		progress(opts, fmt.Sprintf("⊘ Skipped %d external deps", len(extResult.Skipped)))
	}

	return nil
}

// configureMachine configures machine-specific settings
func configureMachine(cfg *config.Config, opts InstallOptions, result *InstallResult) error {
	if len(cfg.MachineConfig) == 0 {
		return nil
	}

	progress(opts, "\n── Machine Configuration ──")

	// Check which configs are missing
	statuses := machine.CheckMachineConfigStatus(cfg)
	var needsConfig []config.MachinePrompt

	for _, status := range statuses {
		if status.Status == "missing" {
			mc := machine.GetMachineConfigByID(cfg, status.ID)
			if mc != nil {
				needsConfig = append(needsConfig, *mc)
			}
		}
	}

	if len(needsConfig) == 0 {
		progress(opts, "✓ All machine configs are already set up")
		return nil
	}

	progress(opts, fmt.Sprintf("Configuring %d machine settings...", len(needsConfig)))

	promptOpts := machine.PromptOptions{
		SkipPrompts: opts.Auto,
		ProgressFunc: func(msg string) {
			progress(opts, "  "+msg)
		},
	}

	renderOpts := machine.RenderOptions{
		Overwrite: opts.Overwrite,
		ProgressFunc: func(msg string) {
			progress(opts, "  "+msg)
		},
	}

	// Collect and render each config
	for _, mc := range needsConfig {
		promptResult, err := machine.CollectSingleConfig(cfg, mc.ID, promptOpts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to collect %s: %w", mc.ID, err))
			continue
		}

		renderResult, err := machine.RenderAndWrite(&mc, promptResult.Values, renderOpts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to write %s: %w", mc.ID, err))
			continue
		}

		result.MachineConfigs = append(result.MachineConfigs, *renderResult)
	}

	if len(result.MachineConfigs) > 0 {
		progress(opts, fmt.Sprintf("✓ Configured %d machine settings", len(result.MachineConfigs)))
	}

	return nil
}

// progress sends a progress message if the callback is set
func progress(opts InstallOptions, msg string) {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(msg)
	}
}

// Summary returns a human-readable summary of the installation result
func (r *InstallResult) Summary() string {
	var summary string

	summary += fmt.Sprintf("Platform: %s", r.Platform.OS)
	if r.Platform.Distro != "" {
		summary += fmt.Sprintf(" (%s)", r.Platform.Distro)
	}
	summary += "\n"

	if len(r.DepsInstalled) > 0 || len(r.DepsFailed) > 0 {
		summary += fmt.Sprintf("Dependencies: %d installed, %d failed\n",
			len(r.DepsInstalled), len(r.DepsFailed))
	}

	if len(r.ConfigsStowed) > 0 || len(r.ConfigsFailed) > 0 {
		summary += fmt.Sprintf("Configs: %d stowed, %d failed\n",
			len(r.ConfigsStowed), len(r.ConfigsFailed))
	}

	if len(r.ExternalCloned) > 0 || len(r.ExternalFailed) > 0 {
		summary += fmt.Sprintf("External: %d cloned, %d failed\n",
			len(r.ExternalCloned), len(r.ExternalFailed))
	}

	if len(r.MachineConfigs) > 0 {
		summary += fmt.Sprintf("Machine configs: %d configured\n", len(r.MachineConfigs))
	}

	return summary
}
