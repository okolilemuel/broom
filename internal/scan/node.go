package scan

import (
	"context"
	"path/filepath"

	"broom/internal/platform"
)

const minNodeSize = 10 * 1024 * 1024 // 10 MB

type NodeScanner struct{}

func (s *NodeScanner) Name() string { return "node_modules" }

func (s *NodeScanner) Scan(ctx context.Context, roots []platform.Root, out chan<- Item) {
	home := homeDir()
	seen := map[string]bool{}

	for _, root := range roots {
		WalkDirs(ctx, root.Path, "node_modules", root.MaxDepth, func(path string) {
			if seen[path] {
				return
			}
			// skip nested node_modules inside node_modules
			if NestedCount(path, "node_modules") > 1 {
				return
			}
			seen[path] = true

			size := DirSizeBytes(path)
			if size < minNodeSize {
				return
			}

			parentDir := filepath.Dir(path)
			lastCommit, hasGit := GitAge(parentDir)
			displayName := ShortPath(home, parentDir)

			out <- Item{
				Path:        path,
				DisplayName: displayName,
				Category:    CategoryNodeModules,
				SizeBytes:   size,
				LastCommit:  lastCommit,
				HasGit:      hasGit,
				Description: "node_modules",
			}
		})
	}
}
