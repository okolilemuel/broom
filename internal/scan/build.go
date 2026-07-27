package scan

import (
	"context"
	"path/filepath"

	"broom/internal/platform"
)

const minBuildSize = 20 * 1024 * 1024 // 20 MB

// buildTargets are common build output directory names.
var buildTargets = []string{".next", "dist", "build", ".turbo", ".parcel-cache", "out", ".output"}

type BuildScanner struct{}

func (s *BuildScanner) Name() string { return "build output" }

func (s *BuildScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	home := homeDir()
	seen := map[string]bool{}

	for _, root := range roots {
		for _, target := range buildTargets {
			WalkDirs(ctx, root.Path, target, root.MaxDepth, func(path string) {
				if seen[path] {
					return
				}
				// skip build dirs inside node_modules
				if NestedCount(path, "node_modules") > 0 {
					return
				}
				seen[path] = true

				size := DirSizeBytes(path)
				if size < minBuildSize {
					return
				}

				parentDir := filepath.Dir(path)
				lastCommit, hasGit := GitAge(parentDir)
				displayName := ShortPath(home, parentDir)

				out <- Item{
					Path:        path,
					DisplayName: displayName,
					Category:    CategoryBuildOutput,
					SizeBytes:   size,
					LastCommit:  lastCommit,
					HasGit:      hasGit,
					Description: filepath.Base(path) + "/",
				}
			})
		}
	}
}
