package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"broom/internal/ui"
)

func main() {
	dryRun := false
	for _, arg := range os.Args[1:] {
		if arg == "--dry-run" || arg == "-n" {
			dryRun = true
		}
		if arg == "--help" || arg == "-h" {
			fmt.Println("Usage: broom [--dry-run]")
			fmt.Println()
			fmt.Println("  --dry-run, -n   Scan and select items but don't delete anything")
			os.Exit(0)
		}
	}

	p := tea.NewProgram(ui.New(dryRun), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "broom: %v\n", err)
		os.Exit(1)
	}
}
