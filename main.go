package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"broom/internal/ui"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func main() {
	dryRun := false
	var olderThan time.Duration

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run" || arg == "-n":
			dryRun = true

		case arg == "--version" || arg == "-v":
			fmt.Printf("broom %s\n", version)
			os.Exit(0)

		case arg == "--older-than":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "broom: --older-than requires a value (e.g. 6m, 1y, 90d)")
				os.Exit(1)
			}
			d, err := parseDuration(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "broom: invalid duration %q: %v\n", args[i], err)
				os.Exit(1)
			}
			olderThan = d

		case strings.HasPrefix(arg, "--older-than="):
			val := strings.TrimPrefix(arg, "--older-than=")
			d, err := parseDuration(val)
			if err != nil {
				fmt.Fprintf(os.Stderr, "broom: invalid duration %q: %v\n", val, err)
				os.Exit(1)
			}
			olderThan = d

		case arg == "--help" || arg == "-h":
			fmt.Println("Usage: broom [options]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --dry-run, -n              Scan and select items but don't delete anything")
			fmt.Println("  --older-than <duration>    Auto-select items older than duration (e.g. 6m, 1y, 90d, 180d)")
			fmt.Println("  --version, -v              Print version and exit")
			fmt.Println("  --help, -h                 Show this help")
			os.Exit(0)
		}
	}

	p := tea.NewProgram(ui.New(dryRun, olderThan), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "broom: %v\n", err)
		os.Exit(1)
	}
}

// parseDuration parses broom duration strings: 6m (months), 1y (years), 90d (days).
// It also accepts standard Go durations.
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}

	last := s[len(s)-1]
	num := s[:len(s)-1]

	switch last {
	case 'y', 'Y':
		var n int
		if _, err := fmt.Sscanf(num, "%d", &n); err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid year value %q", num)
		}
		return time.Duration(n) * 365 * 24 * time.Hour, nil

	case 'm', 'M':
		var n int
		if _, err := fmt.Sscanf(num, "%d", &n); err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid month value %q", num)
		}
		return time.Duration(n) * 30 * 24 * time.Hour, nil

	case 'd', 'D':
		var n int
		if _, err := fmt.Sscanf(num, "%d", &n); err != nil || n <= 0 {
			return 0, fmt.Errorf("invalid day value %q", num)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}

	// fall back to standard Go duration parsing (e.g. "720h")
	return time.ParseDuration(s)
}
