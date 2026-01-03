package setup

import (
	"fmt"
	"testing"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/deps"
	"github.com/nvandessel/go4dot/internal/machine"
	"github.com/nvandessel/go4dot/internal/platform"
	"github.com/nvandessel/go4dot/internal/stow"
)

func TestInstallOptionsDefaults(t *testing.T) {
	opts := InstallOptions{}

	if opts.Auto {
		t.Error("Auto should default to false")
	}
	if opts.Minimal {
		t.Error("Minimal should default to false")
	}
	if opts.SkipDeps {
		t.Error("SkipDeps should default to false")
	}
	if opts.SkipExternal {
		t.Error("SkipExternal should default to false")
	}
	if opts.SkipMachine {
		t.Error("SkipMachine should default to false")
	}
	if opts.SkipStow {
		t.Error("SkipStow should default to false")
	}
}

func TestInstallResultHasErrors(t *testing.T) {
	tests := []struct {
		name     string
		result   InstallResult
		hasError bool
	}{
		{
			name:     "No errors",
			result:   InstallResult{},
			hasError: false,
		},
		{
			name: "Has dep failures",
			result: InstallResult{
				DepsFailed: []deps.InstallError{
					{Item: config.DependencyItem{Name: "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "Has config failures",
			result: InstallResult{
				ConfigsFailed: []stow.StowError{
					{ConfigName: "test"},
				},
			},
			hasError: true,
		},
		{
			name: "Has external failures",
			result: InstallResult{
				ExternalFailed: []deps.ExternalError{
					{Dep: config.ExternalDep{Name: "test"}},
				},
			},
			hasError: true,
		},
		{
			name: "Has general errors",
			result: InstallResult{
				Errors: []error{fmt.Errorf("test error")},
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result.HasErrors() != tt.hasError {
				t.Errorf("HasErrors() = %v, want %v", tt.result.HasErrors(), tt.hasError)
			}
		})
	}
}

func TestInstallResultSummary(t *testing.T) {
	result := &InstallResult{
		Platform: &platform.Platform{
			OS:     "linux",
			Distro: "fedora",
		},
		DepsInstalled:  []config.DependencyItem{{Name: "git"}, {Name: "stow"}},
		ConfigsStowed:  []string{"git", "nvim"},
		ExternalCloned: []config.ExternalDep{{Name: "pure"}},
		MachineConfigs: []machine.RenderResult{{ID: "git"}},
	}

	summary := result.Summary()

	if summary == "" {
		t.Error("Summary should not be empty")
	}

	// Check that summary contains expected elements
	if !contains(summary, "linux") {
		t.Error("Summary should contain platform OS")
	}
	if !contains(summary, "fedora") {
		t.Error("Summary should contain distro")
	}
	if !contains(summary, "2 installed") {
		t.Error("Summary should contain deps count")
	}
	if !contains(summary, "2 stowed") {
		t.Error("Summary should contain configs count")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInstallWithSkipAll(t *testing.T) {
	cfg := &config.Config{
		SchemaVersion: "1.0",
		Metadata: config.Metadata{
			Name: "test-dotfiles",
		},
	}

	tmpDir := t.TempDir()

	var progressMessages []string
	opts := InstallOptions{
		SkipDeps:     true,
		SkipStow:     true,
		SkipExternal: true,
		SkipMachine:  true,
		ProgressFunc: func(msg string) {
			progressMessages = append(progressMessages, msg)
		},
	}

	result, err := Install(cfg, tmpDir, opts)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	if result.Platform == nil {
		t.Error("Platform should be detected")
	}

	// With all skips, should have no installed/stowed items
	if len(result.DepsInstalled) != 0 {
		t.Error("Should have no deps installed with skip")
	}
	if len(result.ConfigsStowed) != 0 {
		t.Error("Should have no configs stowed with skip")
	}

	// Should have progress messages for skipping
	if len(progressMessages) == 0 {
		t.Error("Expected progress messages")
	}
}

func TestInstallMinimalMode(t *testing.T) {
	cfg := &config.Config{
		SchemaVersion: "1.0",
		Configs: config.ConfigGroups{
			Core: []config.ConfigItem{
				{Name: "git", Path: "git"},
			},
			Optional: []config.ConfigItem{
				{Name: "nvim", Path: "nvim"},
			},
		},
	}

	tmpDir := t.TempDir()

	opts := InstallOptions{
		Minimal:      true,
		SkipDeps:     true,
		SkipExternal: true,
		SkipMachine:  true,
	}

	result, err := Install(cfg, tmpDir, opts)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// In minimal mode, only core configs should be attempted
	// (they won't stow since dirs don't exist, but that's ok)
	if result.Platform == nil {
		t.Error("Platform should be detected")
	}
}

func TestProgress(t *testing.T) {
	var received string
	opts := InstallOptions{
		ProgressFunc: func(msg string) {
			received = msg
		},
	}

	progress(opts, "test message")

	if received != "test message" {
		t.Errorf("Expected 'test message', got %q", received)
	}
}

func TestProgressNoCallback(t *testing.T) {
	opts := InstallOptions{}

	// Should not panic with nil callback
	progress(opts, "test message")
}
