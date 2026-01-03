package machine

import (
	"bytes"
	"strings"
	"testing"

	"github.com/nvandessel/gopherdot/internal/config"
)

func TestCollectSinglePrompt(t *testing.T) {
	tests := []struct {
		name      string
		prompt    config.PromptField
		input     string
		skipInput bool
		expected  string
		wantErr   bool
	}{
		{
			name: "Simple text input",
			prompt: config.PromptField{
				ID:     "name",
				Prompt: "Enter your name",
				Type:   "text",
			},
			input:    "John Doe\n",
			expected: "John Doe",
		},
		{
			name: "Default value used when empty",
			prompt: config.PromptField{
				ID:      "name",
				Prompt:  "Enter your name",
				Type:    "text",
				Default: "Anonymous",
			},
			input:    "\n",
			expected: "Anonymous",
		},
		{
			name: "Skip prompts uses default",
			prompt: config.PromptField{
				ID:      "name",
				Prompt:  "Enter your name",
				Type:    "text",
				Default: "Default Name",
			},
			skipInput: true,
			expected:  "Default Name",
		},
		{
			name: "Skip prompts fails without default for required",
			prompt: config.PromptField{
				ID:       "name",
				Prompt:   "Enter your name",
				Type:     "text",
				Required: true,
			},
			skipInput: true,
			wantErr:   true,
		},
		{
			name: "Confirm yes",
			prompt: config.PromptField{
				ID:     "enable",
				Prompt: "Enable feature",
				Type:   "confirm",
			},
			input:    "yes\n",
			expected: "true",
		},
		{
			name: "Confirm no",
			prompt: config.PromptField{
				ID:     "enable",
				Prompt: "Enable feature",
				Type:   "confirm",
			},
			input:    "no\n",
			expected: "false",
		},
		{
			name: "Confirm y",
			prompt: config.PromptField{
				ID:     "enable",
				Prompt: "Enable feature",
				Type:   "confirm",
			},
			input:    "y\n",
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in bytes.Buffer
			var out bytes.Buffer

			in.WriteString(tt.input)

			opts := PromptOptions{
				In:          &in,
				Out:         &out,
				SkipPrompts: tt.skipInput,
			}

			reader := strings.NewReader(tt.input)
			bufReader := strings.NewReader(tt.input)

			// For skip prompts test
			if tt.skipInput {
				result, err := collectSinglePrompt(tt.prompt, nil, opts)
				if tt.wantErr {
					if err == nil {
						t.Error("Expected error but got none")
					}
					return
				}
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Got %q, want %q", result, tt.expected)
				}
				return
			}

			// Need to use bufio.Reader
			_ = reader
			_ = bufReader

			// Simplified test using the full collectPrompts flow
			cfg := config.MachinePrompt{
				ID:          "test",
				Description: "Test",
				Prompts:     []config.PromptField{tt.prompt},
				Template:    "",
			}

			result, err := collectPrompts(cfg, opts)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Values[tt.prompt.ID] != tt.expected {
				t.Errorf("Got %q, want %q", result.Values[tt.prompt.ID], tt.expected)
			}
		})
	}
}

func TestCollectMachineConfig(t *testing.T) {
	cfg := &config.Config{
		MachineConfig: []config.MachinePrompt{
			{
				ID:          "git",
				Description: "Git configuration",
				Destination: "~/.gitconfig.local",
				Prompts: []config.PromptField{
					{
						ID:      "user_name",
						Prompt:  "Full name for git commits",
						Type:    "text",
						Default: "Test User",
					},
					{
						ID:      "user_email",
						Prompt:  "Email for git commits",
						Type:    "text",
						Default: "test@example.com",
					},
				},
				Template: "[user]\n    name = {{ .user_name }}\n    email = {{ .user_email }}",
			},
		},
	}

	// Use skip prompts to use defaults
	opts := PromptOptions{
		SkipPrompts: true,
	}

	results, err := CollectMachineConfig(cfg, opts)
	if err != nil {
		t.Fatalf("CollectMachineConfig failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0].ID != "git" {
		t.Errorf("Expected ID 'git', got %q", results[0].ID)
	}

	if results[0].Values["user_name"] != "Test User" {
		t.Errorf("Expected user_name 'Test User', got %q", results[0].Values["user_name"])
	}

	if results[0].Values["user_email"] != "test@example.com" {
		t.Errorf("Expected user_email 'test@example.com', got %q", results[0].Values["user_email"])
	}
}

func TestCollectSingleConfig(t *testing.T) {
	cfg := &config.Config{
		MachineConfig: []config.MachinePrompt{
			{
				ID:          "git",
				Description: "Git configuration",
				Prompts: []config.PromptField{
					{
						ID:      "name",
						Prompt:  "Name",
						Default: "Test",
					},
				},
			},
			{
				ID:          "other",
				Description: "Other config",
				Prompts:     []config.PromptField{},
			},
		},
	}

	opts := PromptOptions{SkipPrompts: true}

	// Test finding existing config
	result, err := CollectSingleConfig(cfg, "git", opts)
	if err != nil {
		t.Fatalf("CollectSingleConfig failed: %v", err)
	}
	if result.ID != "git" {
		t.Errorf("Expected ID 'git', got %q", result.ID)
	}

	// Test not found
	_, err = CollectSingleConfig(cfg, "nonexistent", opts)
	if err == nil {
		t.Error("Expected error for nonexistent config")
	}
}

func TestGetMachineConfigByID(t *testing.T) {
	cfg := &config.Config{
		MachineConfig: []config.MachinePrompt{
			{ID: "git", Description: "Git config"},
			{ID: "ssh", Description: "SSH config"},
		},
	}

	// Test found
	mc := GetMachineConfigByID(cfg, "git")
	if mc == nil {
		t.Fatal("Expected to find 'git' config")
	}
	if mc.ID != "git" {
		t.Errorf("Expected ID 'git', got %q", mc.ID)
	}

	// Test not found
	mc = GetMachineConfigByID(cfg, "nonexistent")
	if mc != nil {
		t.Error("Expected nil for nonexistent config")
	}
}

func TestListMachineConfigs(t *testing.T) {
	cfg := &config.Config{
		MachineConfig: []config.MachinePrompt{
			{ID: "git", Description: "Git config"},
			{ID: "ssh", Description: "SSH config"},
		},
	}

	list := ListMachineConfigs(cfg)

	if len(list) != 2 {
		t.Fatalf("Expected 2 configs, got %d", len(list))
	}

	if list[0].ID != "git" || list[0].Description != "Git config" {
		t.Errorf("Unexpected first item: %+v", list[0])
	}

	if list[1].ID != "ssh" || list[1].Description != "SSH config" {
		t.Errorf("Unexpected second item: %+v", list[1])
	}
}
