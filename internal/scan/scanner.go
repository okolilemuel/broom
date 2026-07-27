package scan

import (
	"context"
	"sync"

	"broom/internal/platform"
)

// ScanContext holds shared state passed to all scanners for the duration of a
// scan run. The Seen map prevents the same real path from being emitted twice
// when multiple search roots overlap (e.g. ~/Projects and ~/).
type ScanContext struct {
	Seen    *sync.Map
	Ignores []string
}

// Scanner finds cleanable items and sends them to out.
type Scanner interface {
	Name() string
	Scan(ctx context.Context, roots []platform.Root, sc ScanContext, out chan<- Item)
}

// Progress reports scanning progress from a single Scanner.
type Progress struct {
	Scanner string
	Found   int
	Done    bool
}

// Run executes all scanners concurrently and streams results.
func Run(
	ctx context.Context,
	scanners []Scanner,
	roots []platform.Root,
	caches []platform.CacheEntry,
	ignores []string,
) (<-chan Item, <-chan Progress) {
	items := make(chan Item, 200)
	progress := make(chan Progress, 100)

	sc := ScanContext{
		Seen:    &sync.Map{},
		Ignores: ignores,
	}

	go func() {
		var wg sync.WaitGroup

		for _, s := range scanners {
			wg.Add(1)
			s := s
			go func() {
				defer wg.Done()
				out := make(chan Item, 50)
				go func() {
					s.Scan(ctx, roots, sc, out)
					close(out)
				}()
				found := 0
				for item := range out {
					items <- item
					found++
					progress <- Progress{Scanner: s.Name(), Found: found}
				}
				progress <- Progress{Scanner: s.Name(), Found: found, Done: true}
			}()
		}

		wg.Wait()
		close(items)
		close(progress)
	}()

	return items, progress
}
