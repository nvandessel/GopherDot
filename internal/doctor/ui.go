package doctor

import (
	"fmt"

	"github.com/nvandessel/go4dot/internal/ui"
)

// PrintReport prints the check result using internal/ui styles.
func PrintReport(result *CheckResult, verbose bool) {
	ui.Section("Health Report")

	// Platform info
	if result.Platform != nil {
		fmt.Printf("Platform: %s", result.Platform.OS)
		if result.Platform.Distro != "" {
			fmt.Printf(" (%s)", result.Platform.Distro)
		}
		fmt.Printf(" [%s]\n\n", result.Platform.PackageManager)
	}

	// Checks
	for _, check := range result.Checks {
		switch check.Status {
		case StatusOK:
			ui.Success("%s: %s", check.Name, check.Message)
		case StatusWarning:
			ui.Warning("%s: %s", check.Name, check.Message)
		case StatusError:
			ui.Error("%s: %s", check.Name, check.Message)
		case StatusSkipped:
			fmt.Printf("  ⊘ %s: %s\n", check.Name, check.Message)
		}

		if verbose && check.Fix != "" && check.Status != StatusOK {
			fmt.Printf("    Fix: %s\n", check.Fix)
		}
	}

	fmt.Println()
	ui.Section("Summary")

	ok, warnings, errors, skipped := result.CountByStatus()
	if errors > 0 {
		ui.Error("%d errors found", errors)
	}
	if warnings > 0 {
		ui.Warning("%d warnings", warnings)
	}
	if ok > 0 {
		ui.Success("%d checks passed", ok)
	}
	if skipped > 0 {
		fmt.Printf("  ⊘ %d skipped\n", skipped)
	}

	// Fixes
	if !result.IsHealthy() || result.HasWarnings() {
		fixes := result.GetFixes()
		if len(fixes) > 0 {
			ui.Section("Suggested Fixes")
			for i, fix := range fixes {
				fmt.Printf("%d. %s\n", i+1, fix)
			}
		}
	}
}
