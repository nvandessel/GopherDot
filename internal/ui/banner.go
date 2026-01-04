package ui

import (
	_ "embed"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

//go:embed banner.txt
var banner string

// PrintBanner prints the ASCII art banner
func PrintBanner(version string) {
	fmt.Println(lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Render(banner))

	fmt.Println(lipgloss.NewStyle().
		Foreground(SubtleColor).
		Render("           v" + version))
	fmt.Println()
}
