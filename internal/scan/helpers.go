package scan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func homeDir() string {
	h, _ := os.UserHomeDir()
	return h
}

// skipAtRoot are directory names to never recurse into from any root.
var skipAtRoot = map[string]bool{
	"Library": true, "Applications": true, ".Trash": true,
	"System": true, "Volumes": true, "proc": true, "sys": true,
	"snap": true, "nix": true, ".cargo": true, ".npm": true,
	".gradle": true, ".m2": true, "go": true,
}

// DedupeKey returns the canonical real path for deduplication.
// It calls filepath.EvalSymlinks to resolve symlinks; falls back to the
// original path if that fails (e.g., the path doesn't exist yet).
func DedupeKey(path string) string {
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return real
}

// WalkDirs walks root looking for directories named target, up to maxDepth.
// It calls fn for each match and does not recurse into the matched dir.
// Paths whose cleaned form has a prefix listed in ignores are skipped.
func WalkDirs(ctx context.Context, root, target string, maxDepth int, ignores []string, fn func(path string)) {
	var walk func(dir string, depth int)
	walk = func(dir string, depth int) {
		if depth > maxDepth {
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			path := filepath.Join(dir, name)

			// check ignore list
			if isIgnored(path, ignores) {
				continue
			}

			if name == target {
				fn(path)
				continue // don't recurse into the found dir
			}
			// skip hidden dirs (except venv-style names like .venv)
			if strings.HasPrefix(name, ".") && name != ".venv" {
				continue
			}
			// skip known non-project dirs at root level
			if depth == 0 && skipAtRoot[name] {
				continue
			}
			walk(path, depth+1)
		}
	}
	walk(root, 0)
}

// isIgnored returns true if path has any of the ignore prefixes.
func isIgnored(path string, ignores []string) bool {
	clean := filepath.Clean(path)
	for _, prefix := range ignores {
		if clean == prefix || strings.HasPrefix(clean, prefix+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// DirSizeBytes returns the size of a directory in bytes using `du`.
func DirSizeBytes(path string) int64 {
	out, err := exec.Command("du", "-sk", path).Output()
	if err != nil {
		return 0
	}
	var kb int64
	fmt.Sscanf(string(out), "%d", &kb)
	return kb * 1024
}

// GitAge returns the last commit time for a directory's git repo.
func GitAge(dir string) (time.Time, bool) {
	out, err := exec.Command("git", "-C", dir, "log", "-1", "--format=%ai").Output()
	if err != nil {
		return time.Time{}, false
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02 15:04:05 -0700", s)
	if err != nil {
		return time.Time{}, true
	}
	return t, true
}

// ShortPath shortens a path relative to a base for display.
func ShortPath(base, path string) string {
	rel, err := filepath.Rel(base, path)
	if err != nil {
		return filepath.Base(path)
	}
	return rel
}

// PathExists returns true if a path exists on disk.
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// NestedCount counts how many times name appears as a path component.
func NestedCount(path, name string) int {
	count := 0
	for _, part := range strings.Split(path, string(os.PathSeparator)) {
		if part == name {
			count++
		}
	}
	return count
}
