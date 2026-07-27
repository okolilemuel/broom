package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadIgnores reads ~/.config/broom/ignore and returns a slice of path prefixes
// to skip during scanning. Lines starting with # are comments. ~ is expanded.
func LoadIgnores() []string {
	home, _ := os.UserHomeDir()
	cfgPath := filepath.Join(home, ".config", "broom", "ignore")

	f, err := os.Open(cfgPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var ignores []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// expand leading ~
		if strings.HasPrefix(line, "~/") {
			line = filepath.Join(home, line[2:])
		} else if line == "~" {
			line = home
		}
		ignores = append(ignores, line)
	}
	return ignores
}
