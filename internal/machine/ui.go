package machine

import (
	"fmt"

	"github.com/nvandessel/go4dot/internal/ui"
)

// PrintStatus prints the machine configuration status using internal/ui styles.
func PrintStatus(statuses []MachineConfigStatus) {
	ui.Section("Machine Configuration Status")

	var configured, missing int
	for _, s := range statuses {
		switch s.Status {
		case "configured":
			ui.Success("%s (%s)", s.Description, s.Destination)
			configured++
		case "missing":
			fmt.Printf("  • %s (not configured)\n", s.Description)
			missing++
		case "error":
			ui.Error("%s: %s", s.Description, s.Error)
		}
	}

	fmt.Println()
	ui.Section("Summary")
	fmt.Printf("Configured: %d\n", configured)
	fmt.Printf("Missing:    %d\n", missing)

	if missing > 0 {
		fmt.Println("\nRun 'g4d machine configure' to set up missing configurations.")
	}
}

// PrintSystemInfo prints the system information using internal/ui styles.
func PrintSystemInfo(info *SystemInfo) {
	ui.Section("System Information")
	fmt.Printf("Username:   %s\n", info.Username)
	fmt.Printf("Hostname:   %s\n", info.Hostname)
	fmt.Printf("Home:       %s\n", info.HomeDir)

	ui.Section("Git Configuration")
	if info.GitUserName != "" {
		fmt.Printf("user.name:  %s\n", info.GitUserName)
	} else {
		fmt.Println("user.name:  (not configured)")
	}
	if info.GitEmail != "" {
		fmt.Printf("user.email: %s\n", info.GitEmail)
	} else {
		fmt.Println("user.email: (not configured)")
	}

	ui.Section("Security Keys")
	if info.HasGPG {
		ui.Success("GPG: Keys available")
		keys, _ := DetectGPGKeys()
		for _, key := range keys {
			fmt.Printf("      - %s (%s)\n", key.Email, key.KeyID)
		}
	} else {
		fmt.Println("  x GPG: No keys found")
	}

	if info.HasSSH {
		ui.Success("SSH: Keys loaded")
		keys, _ := DetectSSHKeys()
		for _, key := range keys {
			fmt.Printf("      - %s (%s)\n", key.Path, key.Type)
		}
	} else {
		fmt.Println("  x SSH: No keys loaded in agent")
	}
}
