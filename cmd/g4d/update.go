package main

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
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update [config-path]",
	Short: "Update dotfiles from git",
	Long: `Pull latest changes from git and update dotfiles.

This command:
1. Runs git pull in the dotfiles directory
2. Shows what files changed
3. Restows all configs to apply changes
4. Updates external dependencies (if --external flag is set)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Load state to get dotfiles path
		st, err := state.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}

		var dotfilesPath string
		var cfg *config.Config

		if len(args) > 0 {
			cfg, err = config.LoadFromPath(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			dotfilesPath = filepath.Dir(args[0])
		} else if st != nil && st.DotfilesPath != "" {
			dotfilesPath = st.DotfilesPath
			cfg, err = config.LoadFromPath(dotfilesPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
		} else {
			cfg, dotfilesPath, err = config.LoadFromDiscovery()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				os.Exit(1)
			}
			dotfilesPath = filepath.Dir(dotfilesPath)
		}

		updateExternal, _ := cmd.Flags().GetBool("external")
		skipRestow, _ := cmd.Flags().GetBool("skip-restow")

		fmt.Println("Updating dotfiles...")
		fmt.Printf("Directory: %s\n\n", dotfilesPath)

		// Check if it's a git repo
		gitDir := filepath.Join(dotfilesPath, ".git")
		if _, err := os.Stat(gitDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: %s is not a git repository\n", dotfilesPath)
			os.Exit(1)
		}

		// Get current HEAD
		oldHead, err := gitHead(dotfilesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not get current HEAD: %v\n", err)
		}

		// Run git pull
		fmt.Println("Pulling latest changes...")
		pullCmd := exec.Command("git", "pull", "--rebase")
		pullCmd.Dir = dotfilesPath
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr

		if err := pullCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: git pull failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println()

		// Get new HEAD
		newHead, err := gitHead(dotfilesPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not get new HEAD: %v\n", err)
		}

		// Show what changed
		if oldHead != "" && newHead != "" && oldHead != newHead {
			fmt.Println("Changes:")
			diffCmd := exec.Command("git", "log", "--oneline", oldHead+".."+newHead)
			diffCmd.Dir = dotfilesPath
			diffCmd.Stdout = os.Stdout
			diffCmd.Stderr = os.Stderr
			diffCmd.Run()
			fmt.Println()

			// Check if config file changed
			configChanged, _ := gitFileChanged(dotfilesPath, oldHead, newHead, config.ConfigFileName)
			if configChanged {
				fmt.Printf("Note: %s was updated. Reloading config...\n\n", config.ConfigFileName)
				cfg, err = config.LoadFromPath(dotfilesPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error reloading config: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			fmt.Println("Already up to date.")
		}

		// Restow configs
		if !skipRestow {
			fmt.Println("Restowing configs...")

			stowOpts := stow.StowOptions{
				ProgressFunc: func(msg string) {
					fmt.Println("  " + msg)
				},
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
					fmt.Printf("\nWarning: %d configs failed to restow:\n", len(result.Failed))
					for _, f := range result.Failed {
						fmt.Printf("  - %s: %v\n", f.ConfigName, f.Error)
					}
				} else {
					fmt.Printf("Restowed %d configs\n", len(result.Success))
				}
			}
			fmt.Println()
		}

		// Update external deps if requested
		if updateExternal && len(cfg.External) > 0 {
			fmt.Println("Updating external dependencies...")

			p, err := platform.Detect()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to detect platform: %v\n", err)
			} else {
				extOpts := deps.ExternalOptions{
					Update:   true,
					RepoRoot: dotfilesPath,
					ProgressFunc: func(msg string) {
						fmt.Println("  " + msg)
					},
				}

				result, err := deps.CloneExternal(cfg, p, extOpts)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to update externals: %v\n", err)
				} else {
					if len(result.Updated) > 0 {
						fmt.Printf("Updated %d external dependencies\n", len(result.Updated))
					}
					if len(result.Failed) > 0 {
						fmt.Printf("Warning: %d external deps failed to update:\n", len(result.Failed))
						for _, f := range result.Failed {
							fmt.Printf("  - %s: %v\n", f.Dep.Name, f.Error)
						}
					}
				}
			}
			fmt.Println()
		}

		// Update state
		if st != nil {
			st.DotfilesPath = dotfilesPath
			if err := st.Save(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
			}
		}

		fmt.Println("Update complete!")
	},
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

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().Bool("external", false, "Also update external dependencies")
	updateCmd.Flags().Bool("skip-restow", false, "Skip restowing configs after pull")
}
