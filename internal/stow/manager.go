package stow

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nvandessel/go4dot/internal/config"
)

// StowResult represents the result of a stow operation
type StowResult struct {
	Success []string // Successfully stowed configs
	Failed  []StowError
	Skipped []string // Skipped (already stowed, conflicts, etc.)
}

// StowError represents a stow operation error
type StowError struct {
	ConfigName string
	Error      error
}

// StowOptions configures stow behavior
type StowOptions struct {
	DryRun       bool
	Force        bool // Overwrite conflicts
	ProgressFunc func(msg string)
}

// Stow symlinks a config directory using GNU stow
func Stow(dotfilesPath string, configName string, opts StowOptions) error {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("Stowing %s...", configName))
	}

	// Build stow command
	args := []string{"-v"} // Verbose

	if opts.DryRun {
		args = append(args, "-n") // No-op/dry-run
	}

	if opts.Force {
		args = append(args, "--adopt") // Adopt existing files
	}

	args = append(args, "-t", os.Getenv("HOME")) // Target home directory
	args = append(args, "-d", dotfilesPath)      // Directory containing packages
	args = append(args, configName)              // Package to stow

	cmd := exec.Command("stow", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("stow failed: %w\nOutput: %s", err, string(output))
	}

	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("✓ Stowed %s", configName))
	}

	return nil
}

// Unstow removes symlinks for a config
func Unstow(dotfilesPath string, configName string, opts StowOptions) error {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("Unstowing %s...", configName))
	}

	args := []string{"-v", "-D"} // Delete/unstow

	if opts.DryRun {
		args = append(args, "-n")
	}

	args = append(args, "-t", os.Getenv("HOME"))
	args = append(args, "-d", dotfilesPath)
	args = append(args, configName)

	cmd := exec.Command("stow", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("unstow failed: %w\nOutput: %s", err, string(output))
	}

	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("✓ Unstowed %s", configName))
	}

	return nil
}

// Restow refreshes symlinks for a config (unstow + stow)
func Restow(dotfilesPath string, configName string, opts StowOptions) error {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("Restowing %s...", configName))
	}

	args := []string{"-v", "-R"} // Restow

	if opts.DryRun {
		args = append(args, "-n")
	}

	if opts.Force {
		args = append(args, "--adopt")
	}

	args = append(args, "-t", os.Getenv("HOME"))
	args = append(args, "-d", dotfilesPath)
	args = append(args, configName)

	cmd := exec.Command("stow", args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("restow failed: %w\nOutput: %s", err, string(output))
	}

	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("✓ Restowed %s", configName))
	}

	return nil
}

// StowConfigs stows multiple configs
func StowConfigs(dotfilesPath string, configs []config.ConfigItem, opts StowOptions) *StowResult {
	result := &StowResult{}

	for _, cfg := range configs {
		// Check if config directory exists
		configPath := filepath.Join(dotfilesPath, cfg.Path)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			result.Skipped = append(result.Skipped, cfg.Name)
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(fmt.Sprintf("⊘ Skipped %s (directory not found)", cfg.Name))
			}
			continue
		}

		// Stow it
		err := Stow(dotfilesPath, cfg.Path, opts)
		if err != nil {
			result.Failed = append(result.Failed, StowError{
				ConfigName: cfg.Name,
				Error:      err,
			})
		} else {
			result.Success = append(result.Success, cfg.Name)
		}
	}

	return result
}

// UnstowConfigs unstows multiple configs
func UnstowConfigs(dotfilesPath string, configs []config.ConfigItem, opts StowOptions) *StowResult {
	result := &StowResult{}

	for _, cfg := range configs {
		err := Unstow(dotfilesPath, cfg.Path, opts)
		if err != nil {
			result.Failed = append(result.Failed, StowError{
				ConfigName: cfg.Name,
				Error:      err,
			})
		} else {
			result.Success = append(result.Success, cfg.Name)
		}
	}

	return result
}

// RestowConfigs restows multiple configs
func RestowConfigs(dotfilesPath string, configs []config.ConfigItem, opts StowOptions) *StowResult {
	result := &StowResult{}

	for _, cfg := range configs {
		configPath := filepath.Join(dotfilesPath, cfg.Path)
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			result.Skipped = append(result.Skipped, cfg.Name)
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(fmt.Sprintf("⊘ Skipped %s (directory not found)", cfg.Name))
			}
			continue
		}

		err := Restow(dotfilesPath, cfg.Path, opts)
		if err != nil {
			result.Failed = append(result.Failed, StowError{
				ConfigName: cfg.Name,
				Error:      err,
			})
		} else {
			result.Success = append(result.Success, cfg.Name)
		}
	}

	return result
}

// IsStowInstalled checks if GNU stow is available
func IsStowInstalled() bool {
	_, err := exec.LookPath("stow")
	return err == nil
}

// ValidateStow checks if stow is installed and working
func ValidateStow() error {
	if !IsStowInstalled() {
		return fmt.Errorf("GNU stow is not installed")
	}

	// Try to get stow version
	cmd := exec.Command("stow", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("stow command failed: %w", err)
	}

	// Check if it's actually GNU stow
	if !strings.Contains(string(output), "GNU Stow") && !strings.Contains(string(output), "stow") {
		return fmt.Errorf("unexpected stow version output: %s", string(output))
	}

	return nil
}
