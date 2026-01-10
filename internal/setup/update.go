package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/deps"
	"github.com/nvandessel/go4dot/internal/platform"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/nvandessel/go4dot/internal/stow"
)

// UpdateOptions configures the update behavior.
type UpdateOptions struct {
	UpdateExternal bool
	SkipRestow     bool
	ProgressFunc   func(current, total int, msg string)
}

// Update pulls latest changes from git and updates dotfiles.
func Update(cfg *config.Config, dotfilesPath string, st *state.State, opts UpdateOptions) error {
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(0, 0, fmt.Sprintf("Updating dotfiles in %s...", dotfilesPath))
	}

	// Check if it's a git repo
	gitDir := filepath.Join(dotfilesPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("%s is not a git repository", dotfilesPath)
	}

	// Get current HEAD
	oldHead, err := gitHead(dotfilesPath)
	if err != nil {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: could not get current HEAD: %v", err))
		}
	}

	// Run git pull
	if opts.ProgressFunc != nil {
		opts.ProgressFunc(0, 0, "Pulling latest changes...")
	}
	pullCmd := exec.Command("git", "pull", "--rebase")
	pullCmd.Dir = dotfilesPath
	if output, err := pullCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %w\nOutput: %s", err, string(output))
	}

	// Get new HEAD
	newHead, err := gitHead(dotfilesPath)
	if err != nil {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: could not get new HEAD: %v", err))
		}
	}

	// Show what changed
	if oldHead != "" && newHead != "" && oldHead != newHead {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Changes detected. Reloading config if needed...")
		}

		// Check if config file changed
		configChanged, _ := gitFileChanged(dotfilesPath, oldHead, newHead, config.ConfigFileName)
		if configChanged {
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(0, 0, fmt.Sprintf("  Note: %s was updated. Reloading config...", config.ConfigFileName))
			}
			newCfg, err := config.LoadFromPath(dotfilesPath)
			if err == nil {
				*cfg = *newCfg
			} else {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: failed to reload config: %v", err))
				}
			}
		}
	} else {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Already up to date.")
		}
	}

	// Restow configs
	if !opts.SkipRestow {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Restowing configs...")
		}

		stowOpts := stow.StowOptions{
			ProgressFunc: opts.ProgressFunc,
		}

		// Get configs to restow (from state or all from config)
		var configsToRestow []config.ConfigItem
		if st != nil && len(st.Configs) > 0 {
			// Restow only installed configs
			for _, sc := range st.Configs {
				if item := cfg.GetConfigByName(sc.Name); item != nil {
					configsToRestow = append(configsToRestow, *item)
				}
			}
		} else {
			// Restow all core configs
			configsToRestow = cfg.Configs.Core
		}

		if len(configsToRestow) > 0 {
			result := stow.RestowConfigs(dotfilesPath, configsToRestow, stowOpts)

			if len(result.Failed) > 0 {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ %d configs failed to restow", len(result.Failed)))
				}
			} else {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("✓ Restowed %d configs", len(result.Success)))
				}
			}
		}
	}

	// Update external deps if requested
	if opts.UpdateExternal && len(cfg.External) > 0 {
		if opts.ProgressFunc != nil {
			opts.ProgressFunc(0, 0, "Updating external dependencies...")
		}

		p, err := platform.Detect()
		if err != nil {
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: failed to detect platform: %v", err))
			}
		} else {
			extOpts := deps.ExternalOptions{
				Update:       true,
				RepoRoot:     dotfilesPath,
				ProgressFunc: opts.ProgressFunc,
			}

			result, err := deps.CloneExternal(cfg, p, extOpts)
			if err != nil {
				if opts.ProgressFunc != nil {
					opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: failed to update externals: %v", err))
				}
			} else {
				if len(result.Updated) > 0 {
					if opts.ProgressFunc != nil {
						opts.ProgressFunc(0, 0, fmt.Sprintf("✓ Updated %d external dependencies", len(result.Updated)))
					}
				}
				if len(result.Failed) > 0 {
					if opts.ProgressFunc != nil {
						opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ %d external deps failed to update", len(result.Failed)))
					}
				}
			}
		}
	}

	// Update state
	if st != nil {
		st.DotfilesPath = dotfilesPath
		if err := st.Save(); err != nil {
			if opts.ProgressFunc != nil {
				opts.ProgressFunc(0, 0, fmt.Sprintf("  ⚠ Warning: failed to save state: %v", err))
			}
		}
	}

	return nil
}

// gitHead returns the current HEAD commit hash
func gitHead(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// gitFileChanged checks if a file changed between two commits
func gitFileChanged(dir, oldCommit, newCommit, filename string) (bool, error) {
	cmd := exec.Command("git", "diff", "--name-only", oldCommit, newCommit, "--", filename)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}
