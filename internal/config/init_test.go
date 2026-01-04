package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfig(t *testing.T) {
	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "g4d-init-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create dummy dirs
	if err := os.Mkdir(filepath.Join(tmpDir, "nvim"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "zsh"), 0755); err != nil {
		t.Fatal(err)
	}

	// Prepare input
	// 1. Name: TestProj
	// 2. Author: Tester
	// 3. Desc: (default)
	// 4. Repo: (default)
	// 5. Add nvim? y
	// 6. Desc nvim: (default)
	// 7. Add zsh? n
	input := "TestProj\nTester\n\n\ny\n\nn\n"
	in := strings.NewReader(input)
	var out bytes.Buffer

	err = InitConfigWithIO(tmpDir, in, &out)
	if err != nil {
		t.Fatalf("InitConfigWithIO failed: %v", err)
	}

	// Check output
	output := out.String()
	if !strings.Contains(output, "Scanning") {
		t.Error("Output should contain 'Scanning'")
	}

	// Check file created
	configPath := filepath.Join(tmpDir, ".go4dot.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Metadata.Name != "TestProj" {
		t.Errorf("Expected name 'TestProj', got '%s'", cfg.Metadata.Name)
	}
	
	// Check configs
	// We expect 1 config (nvim) because zsh was rejected
	if len(cfg.Configs.Core) != 1 {
		t.Errorf("Expected 1 config, got %d", len(cfg.Configs.Core))
	}
	if len(cfg.Configs.Core) > 0 && cfg.Configs.Core[0].Name != "nvim" {
		t.Errorf("Expected nvim, got %s", cfg.Configs.Core[0].Name)
	}
}
