package scan

import (
	"context"
	"path/filepath"

	"broom/internal/platform"
)

const minRustTargetSize = 50 * 1024 * 1024 // 50 MB

type RustScanner struct{}

func (s *RustScanner) Name() string { return "rust artifacts" }

func (s *RustScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	home := homeDir()
	seen := map[string]bool{}

	for _, root := range roots {
		WalkDirs(ctx, root.Path, "target", root.MaxDepth, func(path string) {
			if seen[path] {
				return
			}
			parentDir := filepath.Dir(path)
			// confirm it's a Rust project by checking for Cargo.toml nearby
			if !PathExists(filepath.Join(parentDir, "Cargo.toml")) &&
				!PathExists(filepath.Join(filepath.Dir(parentDir), "Cargo.toml")) {
				return
			}
			seen[path] = true

			size := DirSizeBytes(path)
			if size < minRustTargetSize {
				return
			}

			lastCommit, hasGit := GitAge(parentDir)
			displayName := ShortPath(home, parentDir)

			out <- Item{
				Path:        path,
				DisplayName: displayName,
				Category:    CategoryRustArtifact,
				SizeBytes:   size,
				LastCommit:  lastCommit,
				HasGit:      hasGit,
				Description: "target/",
			}
		})
	}
}
