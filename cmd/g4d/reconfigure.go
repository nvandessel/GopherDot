package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nvandessel/go4dot/internal/config"
	"github.com/nvandessel/go4dot/internal/machine"
	"github.com/nvandessel/go4dot/internal/state"
	"github.com/spf13/cobra"
)

var reconfigureCmd = &cobra.Command{
	Use:     "reconfigure [id] [config-path]",
	Aliases: []string{"reconfig"},
	Short:   "Re-run machine-specific configuration",
	Long: `Re-run machine-specific prompts to update configuration.

Without arguments, reconfigures all machine settings.
With an ID argument, reconfigures only that specific setting.

This is useful when:
- You want to change git user/email
- GPG or SSH keys have changed
- You need to update machine-specific paths`,
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

		// Load state
		st, err := state.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load state: %v\n", err)
		}

		overwrite := true // Always overwrite when reconfiguring
		skipPrompts, _ := cmd.Flags().GetBool("defaults")

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
			// Reconfigure single
			mc := machine.GetMachineConfigByID(cfg, specificID)
			if mc == nil {
				fmt.Fprintf(os.Stderr, "Error: machine config '%s' not found\n", specificID)
				os.Exit(1)
			}

			fmt.Printf("Reconfiguring %s...\n\n", mc.Description)

			result, err := machine.CollectSingleConfig(cfg, specificID, promptOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			renderResult, err := machine.RenderAndWrite(mc, result.Values, renderOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Update state
			if st != nil {
				st.SetMachineConfig(specificID, renderResult.Destination, false, false)
				if err := st.Save(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
				}
			}

			fmt.Printf("\nReconfigured: %s\n", renderResult.Destination)
		} else {
			// Reconfigure all
			fmt.Printf("Reconfiguring %d machine settings...\n\n", len(cfg.MachineConfig))

			results, err := machine.CollectMachineConfig(cfg, promptOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			renderResults, err := machine.RenderAll(cfg, results, renderOpts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			// Update state
			if st != nil {
				for _, r := range renderResults {
					st.SetMachineConfig(r.ID, r.Destination, false, false)
				}
				if err := st.Save(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to save state: %v\n", err)
				}
			}

			fmt.Printf("\nReconfigured %d machine settings\n", len(renderResults))
		}
	},
}

func init() {
	rootCmd.AddCommand(reconfigureCmd)

	reconfigureCmd.Flags().Bool("defaults", false, "Use default values without prompting")
}
