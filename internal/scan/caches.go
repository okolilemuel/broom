package scan

import (
	"context"

	"broom/internal/platform"
)

const minCacheSize = 5 * 1024 * 1024 // 5 MB

// CacheScanner checks known global cache directories.
type CacheScanner struct {
	Caches []platform.CacheEntry
}

func (s *CacheScanner) Name() string { return "caches" }

func (s *CacheScanner) Scan(ctx context.Context, roots []platform.Root, sc ScanContext, out chan<- Item) {
	for _, c := range s.Caches {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if !PathExists(c.Path) {
			continue
		}

		// global deduplication via real path
		key := DedupeKey(c.Path)
		if _, loaded := sc.Seen.LoadOrStore(key, true); loaded {
			continue
		}

		size := DirSizeBytes(c.Path)
		if size < minCacheSize {
			continue
		}
		cat := Category(c.Category)
		out <- Item{
			Path:        c.Path,
			DisplayName: c.Name,
			Category:    cat,
			SizeBytes:   size,
			HasGit:      false,
			Description: "global cache",
		}
	}
}
