package scan

import (
	"context"
	"path/filepath"

	"broom/internal/platform"
)

const minVenvSize = 10 * 1024 * 1024 // 10 MB

type PythonScanner struct{}

func (s *PythonScanner) Name() string { return "python venvs" }

func (s *PythonScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	home := homeDir()
	seen := map[string]bool{}
	targets := []string{".venv", "venv", "env"}

	for _, root := range roots {
		for _, target := range targets {
			WalkDirs(ctx, root.Path, target, root.MaxDepth, func(path string) {
				if seen[path] {
					return
				}
				// make sure it's actually a Python venv
				if !PathExists(filepath.Join(path, "bin", "python")) &&
					!PathExists(filepath.Join(path, "bin", "python3")) &&
					!PathExists(filepath.Join(path, "Scripts", "python.exe")) {
					return
				}
				seen[path] = true

				size := DirSizeBytes(path)
				if size < minVenvSize {
					return
				}

				parentDir := filepath.Dir(path)
				lastCommit, hasGit := GitAge(parentDir)
				displayName := ShortPath(home, parentDir)

				out <- Item{
					Path:        path,
					DisplayName: displayName,
					Category:    CategoryPythonVenv,
					SizeBytes:   size,
					LastCommit:  lastCommit,
					HasGit:      hasGit,
					Description: target,
				}
			})
		}
	}
}
