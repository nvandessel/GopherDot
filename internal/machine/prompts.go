package machine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nvandessel/go4dot/internal/config"
)

// PromptResult holds the collected values from prompts
type PromptResult struct {
	ID     string
	Values map[string]string
}

// PromptOptions configures prompt behavior
type PromptOptions struct {
	In           io.Reader        // Input source (defaults to os.Stdin)
	Out          io.Writer        // Output destination (defaults to os.Stdout)
	ProgressFunc func(msg string) // Called for progress updates
	SkipPrompts  bool             // Use defaults without prompting
}

// CollectMachineConfig prompts the user for all machine-specific values
func CollectMachineConfig(cfg *config.Config, opts PromptOptions) ([]PromptResult, error) {
	if opts.In == nil {
		opts.In = os.Stdin
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	var results []PromptResult

	for _, mc := range cfg.MachineConfig {
		result, err := collectPrompts(mc, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to collect prompts for %s: %w", mc.ID, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// CollectSingleConfig prompts for a single machine config by ID
func CollectSingleConfig(cfg *config.Config, id string, opts PromptOptions) (*PromptResult, error) {
	if opts.In == nil {
		opts.In = os.Stdin
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	var found *config.MachinePrompt
	for i := range cfg.MachineConfig {
		if cfg.MachineConfig[i].ID == id {
			found = &cfg.MachineConfig[i]
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("machine config '%s' not found", id)
	}

	result, err := collectPrompts(*found, opts)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// collectPrompts collects values for a single MachinePrompt
func collectPrompts(mc config.MachinePrompt, opts PromptOptions) (PromptResult, error) {
	result := PromptResult{
		ID:     mc.ID,
		Values: make(map[string]string),
	}

	if opts.ProgressFunc != nil {
		opts.ProgressFunc(fmt.Sprintf("Configuring %s...", mc.Description))
	}

	reader := bufio.NewReader(opts.In)

	for _, prompt := range mc.Prompts {
		value, err := collectSinglePrompt(prompt, reader, opts)
		if err != nil {
			return result, err
		}
		result.Values[prompt.ID] = value
	}

	return result, nil
}

// collectSinglePrompt collects a single prompt value
func collectSinglePrompt(prompt config.PromptField, reader *bufio.Reader, opts PromptOptions) (string, error) {
	// If skipping prompts, use default
	if opts.SkipPrompts {
		if prompt.Required && prompt.Default == "" {
			return "", fmt.Errorf("required field '%s' has no default value", prompt.ID)
		}
		return prompt.Default, nil
	}

	// Build prompt string
	promptStr := prompt.Prompt
	if prompt.Default != "" {
		promptStr = fmt.Sprintf("%s [%s]", prompt.Prompt, prompt.Default)
	}
	if prompt.Required {
		promptStr = promptStr + " (required)"
	}
	promptStr = promptStr + ": "

	for {
		fmt.Fprint(opts.Out, promptStr)

		var input string
		var err error

		switch prompt.Type {
		case "password":
			// For password, we'd ideally hide input, but for now just read normally
			// TODO: Use terminal.ReadPassword or similar
			input, err = reader.ReadString('\n')
		case "confirm":
			input, err = reader.ReadString('\n')
			if err == nil {
				input = strings.TrimSpace(strings.ToLower(input))
				if input == "y" || input == "yes" || input == "true" || input == "1" {
					input = "true"
				} else if input == "n" || input == "no" || input == "false" || input == "0" || input == "" {
					input = "false"
				} else {
					fmt.Fprintln(opts.Out, "Please enter yes or no")
					continue
				}
			}
		case "select":
			// For select, we'd show options - for now just accept text
			// TODO: Implement proper select with options
			input, err = reader.ReadString('\n')
		default: // "text" or unspecified
			input, err = reader.ReadString('\n')
		}

		if err != nil {
			if err == io.EOF {
				// Use default if available
				if prompt.Default != "" {
					return prompt.Default, nil
				}
				if prompt.Required {
					return "", fmt.Errorf("required field '%s' not provided", prompt.ID)
				}
				return "", nil
			}
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		// Use default if empty
		if input == "" && prompt.Default != "" {
			input = prompt.Default
		}

		// Check required
		if prompt.Required && input == "" {
			fmt.Fprintln(opts.Out, "This field is required. Please enter a value.")
			continue
		}

		return input, nil
	}
}

// GetMachineConfigByID returns a machine config by its ID
func GetMachineConfigByID(cfg *config.Config, id string) *config.MachinePrompt {
	for i := range cfg.MachineConfig {
		if cfg.MachineConfig[i].ID == id {
			return &cfg.MachineConfig[i]
		}
	}
	return nil
}

// ListMachineConfigs returns all machine config IDs and descriptions
func ListMachineConfigs(cfg *config.Config) []struct {
	ID          string
	Description string
} {
	var list []struct {
		ID          string
		Description string
	}
	for _, mc := range cfg.MachineConfig {
		list = append(list, struct {
			ID          string
			Description string
		}{
			ID:          mc.ID,
			Description: mc.Description,
		})
	}
	return list
}
